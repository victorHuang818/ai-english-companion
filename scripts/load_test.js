import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';

// 从环境变量读取主机地址，默认使用本地测试地址
const HOST = __ENV.API_HOST || '127.0.0.1:8889';
const BASE_URL = `http://${HOST}/api/v1`;
const WS_URL = `ws://${HOST}/ws/companion`;

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

  // 3. 动态创建一个练习场景 (Scenario)
  const scenarioPayload = JSON.stringify({
    name: 'Stress Test Scenario',
    description: 'High-throughput system simulation for speaking practice',
  });
  const scenarioRes = http.post(`${BASE_URL}/scenarios/`, scenarioPayload, {
    headers: authHeader,
  });

  check(scenarioRes, {
    'Create Scenario status is 200': (r) => r.status === 200,
  });
  const scenarioData = JSON.parse(scenarioRes.body);
  const scenarioId = scenarioData.id;

  // 4. 创建一个用户背景档案 (User Profile)
  const profilePayload = JSON.stringify({
    english_level: 'intermediate',
    learning_target: 'Stress testing companion API',
  });
  const profileRes = http.post(`${BASE_URL}/user-profiles/`, profilePayload, {
    headers: authHeader,
  });

  check(profileRes, {
    'Create User Profile status is 200': (r) => r.status === 200,
  });
  const profileData = JSON.parse(profileRes.body);
  const userProfileId = profileData.id;

  if (!scenarioId || !userProfileId) {
    throw new Error(`Failed to dynamically seed data. ScenarioId: ${scenarioId}, UserProfileId: ${userProfileId}`);
  }

  console.log(`[k6 Setup] Token, Scenario (${scenarioId}) and User Profile (${userProfileId}) initialized. Target host: ${HOST}`);
  return {
    token: token,
    userProfileId: userProfileId,
    scenarioId: scenarioId,
  };
}

export default function (data) {
  const token = data.token;
  const authHeader = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  // 1. 模拟业务流：发起 HTTP 请求创建一个口语练习会话 (Session)
  const createSessionPayload = JSON.stringify({
    user_profile_id: data.userProfileId,
    scenario_id: data.scenarioId,
  });
  const createRes = http.post(`${BASE_URL}/practices/`, createSessionPayload, {
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
