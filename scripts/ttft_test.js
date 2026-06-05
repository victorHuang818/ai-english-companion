import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';

// 必须在全局作用域（Init Context）加载二进制 PCM 音频文件
const pcmData = open('./welcome.pcm', 'b');

const HOST = __ENV.API_HOST || '127.0.0.1:8889';
const BASE_URL = `http://${HOST}/api/v1`;
const WS_URL = `ws://${HOST}/ws/interview`;

export const options = {
  vus: 1,
  iterations: 1,
};

export function setup() {
  const email = `ttft_test_${Date.now()}@example.com`;
  const username = `ttft_tester_${Date.now()}`;
  const password = 'ttft_password_123';

  // 1. 注册
  const registerPayload = JSON.stringify({ username, email, password });
  const regRes = http.post(`${BASE_URL}/user/register`, registerPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  // 2. 登录
  const loginPayload = JSON.stringify({ email, password });
  const loginRes = http.post(`${BASE_URL}/user/login`, loginPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  const responseBody = JSON.parse(loginRes.body);
  const token = responseBody.token;

  // 3. 创建岗位
  const authHeader = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };
  const jobPayload = JSON.stringify({
    name: 'Go Developer (TTFT Test)',
    description: 'Testing real latency with PCM voice',
  });
  const jobRes = http.post(`${BASE_URL}/job-profiles/`, jobPayload, { headers: authHeader });
  const jobProfileId = JSON.parse(jobRes.body).id;

  // 4. 获取上传URL
  const uploadUrlRes = http.post(`${BASE_URL}/resumes/upload-url`, JSON.stringify({ filename: 'ttft_resume.pdf' }), { headers: authHeader });
  const uploadUrlData = JSON.parse(uploadUrlRes.body);
  
  // 5. 上传 Mock PDF
  http.put(uploadUrlData.upload_url, 'Mock PDF Content', { headers: { 'Content-Type': 'application/pdf' } });

  // 6. 创建简历记录
  const resumeRes = http.post(`${BASE_URL}/resumes/`, JSON.stringify({ object_key: uploadUrlData.object_key }), { headers: authHeader });
  const resumeId = JSON.parse(resumeRes.body).id;

  return { token, resumeId, jobProfileId };
}

export default function (data) {
  const token = data.token;
  const authHeader = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };

  // 1. 创建面试会话
  const createSessionPayload = JSON.stringify({
    resume_id: data.resumeId,
    job_profile_id: data.jobProfileId,
  });
  const createRes = http.post(`${BASE_URL}/interviews/`, createSessionPayload, { headers: authHeader });
  const sessionId = JSON.parse(createRes.body).session_id;

  // 2. 建立 WebSocket 连接
  const url = `${WS_URL}?token=${encodeURIComponent(token)}`;
  const params = { headers: { 'X-User-ID': 'ttft-test-user' } };

  let sentTime = 0;
  let firstResponseTime = 0;

  const wsRes = ws.connect(url, params, function (socket) {
    socket.on('open', function () {
      // 激活会话
      socket.send(JSON.stringify({ session_id: sessionId }));

      let audioChunksSent = 0;
      const chunkSize = 1280; // 40ms 的 16kHz, 16bit, 单声道 PCM 刚好是 1280 字节

      // 使用全局 setInterval 每 40ms 发送一个真实的语音数据包
      const intervalId = setInterval(function () {
        const offset = audioChunksSent * chunkSize;
        if (offset + chunkSize <= pcmData.byteLength) {
          const chunk = pcmData.slice(offset, offset + chunkSize);
          socket.sendBinary(chunk);
          audioChunksSent++;
        } else {
          // 发送完毕真实音频文件，开始等待响应并记录时间点
          if (sentTime === 0) {
            sentTime = Date.now();
            console.log(`[TTFT Test] Finished sending welcome.pcm at: ${sentTime}. Waiting for Qwen response...`);
          }
          clearInterval(intervalId);
        }
      }, 40);
    });

    socket.on('message', function (message) {
      try {
        const msg = JSON.parse(message);
        
        // 捕获首包响应并打印 TTFT
        if (msg.serverContent && firstResponseTime === 0 && sentTime > 0) {
          firstResponseTime = Date.now();
          console.log(`\n================== TTFT MEASUREMENT ==================`);
          console.log(`[TTFT Test] First response received!`);
          console.log(`[TTFT Test] Audio finished sending: ${sentTime}`);
          console.log(`[TTFT Test] First token received:   ${firstResponseTime}`);
          console.log(`[TTFT Test] Real-time TTFT:         ${firstResponseTime - sentTime} ms`);
          console.log(`======================================================\n`);
          
          socket.close();
        }
      } catch (err) {
        // 忽略非 JSON 数据
      }
    });

    socket.on('error', function (e) {
      console.error(`[k6 WebSocket Error] session_id: ${sessionId}, error: ${e.error()}`);
    });
  });
}
