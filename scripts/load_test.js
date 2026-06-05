import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';

// 从环境变量读取主机地址，默认使用本地测试地址
const HOST = __ENV.API_HOST || '127.0.0.1:8889';
const BASE_URL = `http://${HOST}/api/v1`;
const WS_URL = `ws://${HOST}/ws/interview`;

export const options = {
  stages: [
    { duration: '1m', target: 1500 },  // 1分钟内线性上升到 1500 个并发用户 (VUs)
    { duration: '3m', target: 1500 },  // 持续压测 3 分钟进行高并发压力测试
    { duration: '1m', target: 0 },     // 1分钟内逐渐降开
  ],
};

// Setup 阶段：注册登录账号，并动态创建岗位与上传简历，获取全局共享的 Token 及 IDs
export function setup() {
  const email = `stress_test_${Date.now()}@example.com`;
  const username = `stress_tester_${Date.now()}`;
  const password = 'stress_password_123';

  // 1. 尝试注册测试用户
  const registerPayload = JSON.stringify({
    username: username,
    email: email,
    password: password,
  });
  const regRes = http.post(`${BASE_URL}/user/register`, registerPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(regRes, {
    'Register status is 200': (r) => r.status === 200,
  });

  // 2. 用户登录，获取 JWT Token
  const loginPayload = JSON.stringify({
    email: email,
    password: password,
  });
  const loginRes = http.post(`${BASE_URL}/user/login`, loginPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(loginRes, {
    'Login status is 200': (r) => r.status === 200,
  });

  const responseBody = JSON.parse(loginRes.body);
  const token = responseBody.token;

  if (!token) {
    throw new Error('Failed to obtain JWT token during setup');
  }

  const authHeader = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  // 3. 动态创建一个面试岗位 (Job Profile)
  const jobPayload = JSON.stringify({
    name: 'Go Backend Developer (Stress Test)',
    description: 'High-throughput system simulation',
  });
  const jobRes = http.post(`${BASE_URL}/job-profiles/`, jobPayload, {
    headers: authHeader,
  });

  check(jobRes, {
    'Create Job Profile status is 200': (r) => r.status === 200,
  });
  const jobData = JSON.parse(jobRes.body);
  const jobProfileId = jobData.id;

  // 4. 获取上传简历的预签名 URL
  const uploadUrlPayload = JSON.stringify({
    filename: 'stress_test_resume.pdf',
  });
  const uploadUrlRes = http.post(`${BASE_URL}/resumes/upload-url`, uploadUrlPayload, {
    headers: authHeader,
  });

  check(uploadUrlRes, {
    'Get Upload URL status is 200': (r) => r.status === 200,
  });
  const uploadUrlData = JSON.parse(uploadUrlRes.body);
  const uploadUrl = uploadUrlData.upload_url;
  const objectKey = uploadUrlData.object_key;

  // 5. 上传一个 Mock PDF 文件二进制内容到 MinIO
  const dummyPdf = 'Mock PDF File Content for Stress Testing';
  const putRes = http.put(uploadUrl, dummyPdf, {
    headers: { 'Content-Type': 'application/pdf' },
  });

  check(putRes, {
    'Upload PDF to MinIO status is 200': (r) => r.status === 200,
  });

  // 6. 提交通知，触发简历入库并获取 resume_id
  const resumePayload = JSON.stringify({
    object_key: objectKey,
  });
  const resumeRes = http.post(`${BASE_URL}/resumes/`, resumePayload, {
    headers: authHeader,
  });

  check(resumeRes, {
    'Create Resume status is 200': (r) => r.status === 200,
  });
  const resumeData = JSON.parse(resumeRes.body);
  const resumeId = resumeData.id;

  if (!jobProfileId || !resumeId) {
    throw new Error(`Failed to dynamically seed data. JobProfileId: ${jobProfileId}, ResumeId: ${resumeId}`);
  }

  console.log(`[k6 Setup] Token, Job Profile (${jobProfileId}) and Resume (${resumeId}) initialized. Target host: ${HOST}`);
  return {
    token: token,
    resumeId: resumeId,
    jobProfileId: jobProfileId,
  };
}

export default function (data) {
  const token = data.token;
  const authHeader = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  // 1. 模拟业务流：发起 HTTP 请求创建一个面试会话 (Session)
  const createSessionPayload = JSON.stringify({
    resume_id: data.resumeId,
    job_profile_id: data.jobProfileId,
  });
  const createRes = http.post(`${BASE_URL}/interviews/`, createSessionPayload, {
    headers: authHeader,
  });

  const is200 = check(createRes, {
    'Create Session status is 200': (r) => r.status === 200,
  });

  if (!is200) {
    console.error(`[k6 Error] Create session failed. Status: ${createRes.status}, Body: ${createRes.body}`);
    sleep(1);
    return;
  }

  let sessionData;
  try {
    sessionData = JSON.parse(createRes.body);
  } catch (e) {
    console.error(`[k6 Error] Failed to parse session JSON. Body: ${createRes.body}, Error: ${e}`);
    sleep(1);
    return;
  }

  const sessionId = sessionData.session_id;
  if (!sessionId) {
    sleep(1);
    return;
  }

  // 2. 模拟业务流：发起 WebSocket 连接，并将 token 作为 Query 参数传入
  const url = `${WS_URL}?token=${encodeURIComponent(token)}`;
  const params = { headers: { 'X-User-ID': 'stress-test-user-id' } };

  let sentTime = 0;
  let firstResponseTime = 0;

  const wsRes = ws.connect(url, params, function (socket) {
    socket.on('open', function () {
      // 连接成功，发送当前会话的 session_id 激活面试官
      socket.send(JSON.stringify({ session_id: sessionId }));

      let audioChunksSent = 0;
      
      // 使用全局 setInterval 每 40ms 发送一个 320 字节的假音频包 (对应 16kHz, 16bit, 单声道 PCM 数据)
      const intervalId = setInterval(function () {
        if (audioChunksSent < 50) {
          const dummyAudio = new Uint8Array(320);
          // 如果是进行 TTFT 真实接口测试，填充随机非零噪音以激活大模型的 VAD 活性检测
          if (__ENV.TTFT_TEST === 'true') {
            for (let i = 0; i < 320; i++) {
              dummyAudio[i] = Math.floor(Math.random() * 256);
            }
          }
          socket.sendBinary(dummyAudio.buffer);
          audioChunksSent++;
        } else {
          // 发送完毕 50 包音频（模拟用户说完了约 2 秒的话），取消定时器，静候 AI 响应
          if (sentTime === 0) {
            sentTime = Date.now();
          }
          clearInterval(intervalId);
        }
      }, 40);
    });

    socket.on('message', function (message) {
      if (__ENV.TTFT_TEST === 'true') {
        console.log(`[TTFT Diagnostic] Received raw message: ${message.slice(0, 300)}`);
      }
      try {
        const msg = JSON.parse(message);

        // 捕获首包响应并打印计算出的 TTFT
        if (msg.serverContent && firstResponseTime === 0 && sentTime > 0) {
          firstResponseTime = Date.now();
          console.log(`[TTFT Measure] session_id: ${sessionId} | TTFT: ${firstResponseTime - sentTime} ms`);
          if (__ENV.TTFT_TEST === 'true') {
            console.log(`[TTFT Measure] TTFT test complete. Closing socket...`);
            socket.close();
          }
        }
        
        // 校验接收到的包是否包含合法的实时响应信息
        check(msg, {
          'Received valid response from WS': (m) => m.serverContent !== undefined || m.aiSuggestion !== undefined,
        });

        // 捕获到 AI 面试官说话完毕（本回合结束信号）
        if (msg.serverContent && msg.serverContent.turnComplete) {
          // 模拟人类思考停顿 2 秒
          sleep(2);
          // 挂断结束面试
          socket.close();
        }
      } catch (err) {
        // 捕获解析非 JSON 或其它异常包
      }
    });

    socket.on('error', function (e) {
      console.error(`[k6 WebSocket Error] session_id: ${sessionId}, error: ${e.error()}`);
    });
  });

  check(wsRes, {
    'WebSocket Handshake complete (101)': (r) => r && r.status === 101,
  });

  // 3. 模拟用户完成面试后的停顿（思考时间/思考间歇），避免单个虚拟用户无间隔循环发起新面试
  sleep(1);
}
