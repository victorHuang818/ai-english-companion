import { useAuthStore } from '../store/authStore';

const BASE_URL = '/api/v1';

async function fetchWithAuth(endpoint: string, options: RequestInit = {}) {
  const token = useAuthStore.getState().token;
  
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  } as any;

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    if (response.status === 401) {
      useAuthStore.getState().logout();
      window.location.href = '/auth';
      throw new Error('Session expired. Please log in again.');
    }
    const err = await response.text();
    throw new Error(err || 'API request failed');
  }

  return response.json();
}

export const api = {
  // Auth
  register: (data: any) => fetchWithAuth('/user/register', { method: 'POST', body: JSON.stringify(data) }),
  login: (data: any) => fetchWithAuth('/user/login', { method: 'POST', body: JSON.stringify(data) }),
  getUserInfo: () => fetchWithAuth('/user/info', { method: 'GET' }),
  
  // Resumes
  generateUploadUrl: (filename: string) => fetchWithAuth('/resumes/upload-url', { method: 'POST', body: JSON.stringify({ filename }) }),
  createResume: (objectKey: string) => fetchWithAuth('/resumes/', { method: 'POST', body: JSON.stringify({ object_key: objectKey }) }),
  getResume: (id: string) => fetchWithAuth(`/resumes/${id}`, { method: 'GET' }),
  
  // Jobs
  createJobProfile: (data: any) => fetchWithAuth('/job-profiles/', { method: 'POST', body: JSON.stringify(data) }),
  
  // Interviews
  createSession: (data: any) => fetchWithAuth('/interviews/', { method: 'POST', body: JSON.stringify(data) }),
  listSessions: () => fetchWithAuth('/interviews/', { method: 'GET' }),
  getSessionDetail: (id: string) => fetchWithAuth(`/interviews/${id}`, { method: 'GET' }),
  deleteSession: (id: string) => fetchWithAuth(`/interviews/${id}`, { method: 'DELETE' }),
  getAiSuggestion: (data: any) => fetchWithAuth('/interviews/suggestion', { method: 'POST', body: JSON.stringify(data) }),

  
  // MinIO direct upload
  uploadToMinio: async (url: string, file: File) => {
    // Convert absolute MinIO URL to relative /minio path to go through Nginx/Vite proxy
    const urlObj = new URL(url);
    const uploadUrl = `/minio${urlObj.pathname}${urlObj.search}`;
    const res = await fetch(uploadUrl, {
      method: 'PUT',
      body: file,
      headers: {
        'Content-Type': file.type || 'application/pdf',
      }
    });
    if (!res.ok) throw new Error('Failed to upload file');
  }
};
