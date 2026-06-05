import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { GlassCard } from '../components/GlassCard';
import { GlowingButton } from '../components/GlowingButton';
import { api } from '../services/api';
import { useAuthStore } from '../store/authStore';
import { 
  LogOut, 
  UploadCloud, 
  Briefcase, 
  PlayCircle, 
  History, 
  ExternalLink, 
  Calendar, 
  RefreshCw, 
  Eye, 
  Trash2, 
  Plus, 
  X,
  FileText,
  CheckCircle
} from 'lucide-react';
import { ResumePreviewCard } from '../components/ResumePreviewCard';
import './DashboardPage.css';
import '../components/DashboardModals.css';

interface ResumeHistoryItem {
  id: string;
  filename: string;
  uploadedAt: number;
}

interface JobHistoryItem {
  id: string;
  name: string;
  description: string;
  createdAt: number;
}

export const DashboardPage: React.FC = () => {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();

  // Resume states
  const [resumeFile, setResumeFile] = useState<File | null>(null);
  const [resumeId, setResumeId] = useState<string>('');
  const [resumeData, setResumeData] = useState<string>('');
  const [uploading, setUploading] = useState(false);
  const [isParsing, setIsParsing] = useState(false);
  const [resumeHistory, setResumeHistory] = useState<ResumeHistoryItem[]>([]);
  const [activePreviewResume, setActivePreviewResume] = useState<{ downloadUrl: string; rawText: string; filename: string } | null>(null);

  // Job states
  const [jobName, setJobName] = useState('');
  const [jobDesc, setJobDesc] = useState('');
  const [jobId, setJobId] = useState<string>('');
  const [jobNameInput, setJobNameInput] = useState('');
  const [jobDescInput, setJobDescInput] = useState('');
  const [creatingJob, setCreatingJob] = useState(false);
  const [showJobModal, setShowJobModal] = useState(false);
  const [jobHistory, setJobHistory] = useState<JobHistoryItem[]>([]);
  const [activePreviewJob, setActivePreviewJob] = useState<JobHistoryItem | null>(null);

  // Start Simulation states
  const [showStartModal, setShowStartModal] = useState(false);
  const [selectedStartResumeId, setSelectedStartResumeId] = useState('');
  const [selectedStartJobId, setSelectedStartJobId] = useState('');

  // Simulation & general states
  const [starting, setStarting] = useState(false);
  const [error, setError] = useState('');
  const [sessions, setSessions] = useState<any[]>([]);
  const [loadingSessions, setLoadingSessions] = useState(false);
  const [deleteSessionTargetId, setDeleteSessionTargetId] = useState<string>('');

  // Load history on user load
  useEffect(() => {
    if (user?.id) {
      const savedResumes = localStorage.getItem(`resumes_${user.id}`);
      if (savedResumes) {
        setResumeHistory(JSON.parse(savedResumes));
      }
      const savedJobs = localStorage.getItem(`jobs_${user.id}`);
      if (savedJobs) {
        setJobHistory(JSON.parse(savedJobs));
      }
    }
  }, [user]);

  // Load sessions on mount
  useEffect(() => {
    fetchSessions();
  }, []);

  const fetchSessions = async () => {
    try {
      setLoadingSessions(true);
      const res = await api.listSessions();
      setSessions(res.sessions || []);
    } catch (err: any) {
      console.error('Failed to fetch sessions:', err);
    } finally {
      setLoadingSessions(false);
    }
  };

  const handleLogout = () => {
    logout();
    navigate('/auth');
  };

  // Start polling resume status until it's parsed or failed
  const startPollingResume = (id: string, filename: string) => {
    setIsParsing(true);
    setResumeId(id);
    setResumeData('');
    setError('');

    const interval = setInterval(async () => {
      try {
        const detail = await api.getResume(id);
        if (detail.status === 'parsed') {
          clearInterval(interval);
          setIsParsing(false);
          setResumeData(detail.raw_text);

          // Update local history
          if (user?.id) {
            const savedResumes = localStorage.getItem(`resumes_${user.id}`);
            let currentList: ResumeHistoryItem[] = savedResumes ? JSON.parse(savedResumes) : [];
            if (!currentList.some(item => item.id === id)) {
              currentList.unshift({
                id,
                filename,
                uploadedAt: Date.now()
              });
              localStorage.setItem(`resumes_${user.id}`, JSON.stringify(currentList));
              setResumeHistory(currentList);
            }
          }
        } else if (detail.status === 'failed' || detail.raw_text.includes('Failed')) {
          clearInterval(interval);
          setIsParsing(false);
          setError('Resume parsing failed. Please verify the document format and upload again.');
        }
      } catch (pollErr: any) {
        console.error('Polling error:', pollErr);
      }
    }, 2000);
  };

  const handleUploadResume = async () => {
    if (!resumeFile) return;
    try {
      setUploading(true);
      setError('');
      const { upload_url, object_key } = await api.generateUploadUrl(resumeFile.name);
      await api.uploadToMinio(upload_url, resumeFile);
      const res = await api.createResume(object_key);
      
      startPollingResume(res.id, resumeFile.name);
    } catch (err: any) {
      setError('Failed to upload resume: ' + err.message);
    } finally {
      setUploading(false);
    }
  };

  const handleSelectResume = async (id: string, filename: string) => {
    try {
      setIsParsing(true);
      setResumeId(id);
      setResumeData('');
      setError('');
      const detail = await api.getResume(id);
      if (detail.status === 'parsed') {
        setResumeData(detail.raw_text);
        setIsParsing(false);
      } else {
        startPollingResume(id, filename);
      }
    } catch (err: any) {
      setError('Failed to load selected resume: ' + err.message);
      setIsParsing(false);
    }
  };

  const handlePreviewResume = async (id: string, filename: string) => {
    try {
      setError('');
      const detail = await api.getResume(id);
      setActivePreviewResume({
        downloadUrl: detail.download_url || '',
        rawText: detail.raw_text || '',
        filename: filename
      });
    } catch (err: any) {
      setError('Failed to load preview: ' + err.message);
    }
  };

  const handleDeleteResume = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (user?.id) {
      const filtered = resumeHistory.filter(item => item.id !== id);
      localStorage.setItem(`resumes_${user.id}`, JSON.stringify(filtered));
      setResumeHistory(filtered);
      if (resumeId === id) {
        setResumeId('');
        setResumeData('');
      }
    }
  };

  // Job Profile Handlers
  const handleCreateJob = async () => {
    if (!jobNameInput || !jobDescInput) return;
    try {
      setCreatingJob(true);
      setError('');
      const res = await api.createJobProfile({ name: jobNameInput, description: jobDescInput });
      const newJobId = res.id;
      setJobId(newJobId);
      setJobName(jobNameInput);
      setJobDesc(jobDescInput);

      const newJobItem: JobHistoryItem = {
        id: newJobId,
        name: jobNameInput,
        description: jobDescInput,
        createdAt: Date.now()
      };

      if (user?.id) {
        const savedJobs = localStorage.getItem(`jobs_${user.id}`);
        let currentList: JobHistoryItem[] = savedJobs ? JSON.parse(savedJobs) : [];
        currentList.unshift(newJobItem);
        localStorage.setItem(`jobs_${user.id}`, JSON.stringify(currentList));
        setJobHistory(currentList);
      }

      setShowJobModal(false);
      setJobNameInput('');
      setJobDescInput('');
    } catch (err: any) {
      setError('Failed to create job profile: ' + err.message);
    } finally {
      setCreatingJob(false);
    }
  };

  const handleSelectJob = (job: JobHistoryItem) => {
    setJobId(job.id);
    setJobName(job.name);
    setJobDesc(job.description);
  };

  const handlePreviewJob = (job: JobHistoryItem) => {
    setActivePreviewJob(job);
  };

  const handleDeleteJob = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (user?.id) {
      const filtered = jobHistory.filter(item => item.id !== id);
      localStorage.setItem(`jobs_${user.id}`, JSON.stringify(filtered));
      setJobHistory(filtered);
      if (jobId === id) {
        setJobId('');
        setJobName('');
        setJobDesc('');
      }
    }
  };

  // Start Simulation Handlers
  const handleOpenStartModal = () => {
    setSelectedStartResumeId(resumeId || (resumeHistory.length > 0 ? resumeHistory[0].id : ''));
    setSelectedStartJobId(jobId || (jobHistory.length > 0 ? jobHistory[0].id : ''));
    setShowStartModal(true);
  };

  const handleConfirmStartSession = async () => {
    if (!selectedStartResumeId || !selectedStartJobId) {
      setError('Please select both a resume and a job profile to start the interview.');
      return;
    }
    try {
      setStarting(true);
      setError('');
      const res = await api.createSession({
        resume_id: selectedStartResumeId,
        job_profile_id: selectedStartJobId
      });
      setShowStartModal(false);
      navigate(`/interview/${res.session_id}`);
    } catch (err: any) {
      setError('Failed to start session: ' + err.message);
    } finally {
      setStarting(false);
    }
  };

  const handleConfirmDeleteSession = async () => {
    if (!deleteSessionTargetId) return;
    try {
      setError('');
      await api.deleteSession(deleteSessionTargetId);
      setDeleteSessionTargetId('');
      await fetchSessions();
    } catch (err: any) {
      setError('删除面试记录失败: ' + err.message);
      setDeleteSessionTargetId('');
    }
  };

  return (
    <>
      <div className="dashboard-container animate-fade-in">
        <nav className="dashboard-nav glass-panel">
        <div className="nav-brand">
          <div className="brand-logo">AI</div>
          <span>Interview Matrix</span>
        </div>
        <div className="nav-user">
          <div className="token-badge">
            <span className="token-label">Free Tokens:</span>
            <span className="token-value">{user?.daily_free_tokens || 0}</span>
          </div>
          <div className="user-info">
            <span className="username">{user?.username}</span>
            <button onClick={handleLogout} className="logout-btn" title="Logout">
              <LogOut size={18} />
            </button>
          </div>
        </div>
      </nav>

      {error && <div className="dashboard-error glass-panel">{error}</div>}

      <div className="dashboard-grid">
        {/* Step 1: Identity Matrix (Resume) */}
        <GlassCard className="dashboard-card" glow={!resumeId}>
          <div className="card-header">
            <UploadCloud className="card-icon" />
            <h2>Step 1: Identity Matrix (Resume)</h2>
            {resumeId && !isParsing && (
              <button className="reset-btn" onClick={() => { setResumeId(''); setResumeData(''); }} title="Change Resume">
                <RefreshCw size={14} />
              </button>
            )}
          </div>
          <div className="card-content">
            {isParsing ? (
              <div className="parsing-loader-container">
                <div className="loader-neon-ring"></div>
                <div className="loader-spinner"></div>
                <div className="loader-status-text">Neural Upload Successful</div>
                <div className="loader-sub-text">Parsing competencies via AI processor...</div>
              </div>
            ) : resumeId ? (
              <ResumePreviewCard 
                data={resumeData} 
                onPreviewClick={() => handlePreviewResume(resumeId, resumeHistory.find(r => r.id === resumeId)?.filename || 'Resume')}
              />
            ) : (
              <div className="upload-section">
                <input 
                  type="file" 
                  accept=".pdf" 
                  onChange={e => setResumeFile(e.target.files?.[0] || null)}
                  className="file-input"
                />
                <GlowingButton 
                  variant="secondary" 
                  onClick={handleUploadResume}
                  disabled={!resumeFile || uploading}
                  fullWidth
                >
                  {uploading ? 'Processing...' : 'Upload & Parse PDF'}
                </GlowingButton>

                {resumeHistory.length > 0 && (
                  <div className="history-list-container">
                    <div className="history-list-title">Previously Uploaded Resumes</div>
                    <div className="history-items-scroll">
                      {resumeHistory.map(item => (
                        <div key={item.id} className="history-item" onClick={() => handleSelectResume(item.id, item.filename)}>
                          <FileText size={16} className="item-icon" />
                          <span className="item-filename" title={item.filename}>{item.filename}</span>
                          <div className="item-actions">
                            <button className="action-btn" title="Preview Resume" onClick={(e) => { e.stopPropagation(); handlePreviewResume(item.id, item.filename); }}>
                              <Eye size={14} />
                            </button>
                            <button className="action-btn delete" title="Delete Resume" onClick={(e) => handleDeleteResume(item.id, e)}>
                              <Trash2 size={14} />
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </GlassCard>

        {/* Step 2: Target Parameters (Job) */}
        <GlassCard className="dashboard-card" glow={!!resumeId && !jobId}>
          <div className="card-header">
            <Briefcase className="card-icon" />
            <h2>Step 2: Target Parameters (Job)</h2>
            {jobId && (
              <button className="reset-btn" onClick={() => { setJobId(''); setJobName(''); setJobDesc(''); }} title="Change Job Profile">
                <RefreshCw size={14} />
              </button>
            )}
          </div>
          <div className="card-content">
            {jobId ? (
              <div className="active-job-preview-card animate-fade-in">
                <div className="active-job-header">
                  <CheckCircle size={16} className="active-job-icon" />
                  <span>Target Parameters Configured</span>
                  <button className="preview-action-btn" onClick={() => handlePreviewJob({ id: jobId, name: jobName, description: jobDesc, createdAt: 0 })} title="Preview Job Details">
                    <Eye size={16} />
                    <span>Preview Job</span>
                  </button>
                </div>
                <div className="active-job-body-simple">
                  <h3 className="active-job-title-simple">{jobName}</h3>
                </div>
              </div>
            ) : (
              <div className="upload-section">
                <GlowingButton 
                  variant="secondary" 
                  onClick={() => setShowJobModal(true)}
                  fullWidth
                >
                  <Plus size={18} style={{ marginRight: '8px', verticalAlign: 'middle' }} />
                  Configure Job Profile
                </GlowingButton>

                {jobHistory.length > 0 && (
                  <div className="history-list-container">
                    <div className="history-list-title">Previously Configured Jobs</div>
                    <div className="history-items-scroll">
                      {jobHistory.map(job => (
                        <div key={job.id} className="history-item" onClick={() => handleSelectJob(job)}>
                          <Briefcase size={16} className="item-icon" />
                          <div className="job-info-block">
                            <span className="item-filename" title={job.name}>{job.name}</span>
                          </div>
                          <div className="item-actions">
                            <button className="action-btn" title="Preview Job" onClick={(e) => { e.stopPropagation(); handlePreviewJob(job); }}>
                              <Eye size={14} />
                            </button>
                            <button className="action-btn delete" title="Delete Job" onClick={(e) => handleDeleteJob(job.id, e)}>
                              <Trash2 size={14} />
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </GlassCard>

        {/* Step 3: Enter Simulation */}
        <GlassCard className={`dashboard-card action-card ${(resumeId || resumeHistory.length > 0) && (jobId || jobHistory.length > 0) ? 'ready' : ''}`} glow={!!resumeId && !!jobId}>
          <div className="card-header">
            <PlayCircle className="card-icon" />
            <h2>Step 3: Enter Simulation</h2>
          </div>
          <div className="card-content centered">
            <p className="status-text">
              {(resumeId || resumeHistory.length > 0) && (jobId || jobHistory.length > 0)
                ? 'All systems nominal. Ready for neural link.' 
                : 'Awaiting configuration data...'}
            </p>
            <GlowingButton 
              variant="primary" 
              onClick={handleOpenStartModal}
              disabled={starting || isParsing || (resumeHistory.length === 0 && !resumeId) || (jobHistory.length === 0 && !jobId)}
              className="start-btn"
            >
              {starting ? 'Initializing...' : 'START INTERVIEW'}
            </GlowingButton>
          </div>
        </GlassCard>
      </div>

      {/* Recent Simulations */}
      <div className="history-section">
        <div className="section-header">
          <History size={24} className="accent-icon" />
          <h2>Recent Simulations</h2>
        </div>
        
        <GlassCard className="history-list-card">
          {loadingSessions ? (
            <div className="loading-state">Loading history...</div>
          ) : sessions.length === 0 ? (
            <div className="empty-state">No past simulations found.</div>
          ) : (
            <div className="sessions-list">
              {sessions.map(s => (
                <div key={s.session_id} className="session-row">
                  <div className="session-info">
                    <span className="session-job">{s.job_title}</span>
                    <div className="session-meta">
                      <span className="meta-item"><Calendar size={14} /> {new Date(s.created_at * 1000).toLocaleDateString()}</span>
                      <span className={`status-tag ${s.status.toLowerCase()}`}>{s.status}</span>
                    </div>
                  </div>
                  <div className="session-score">
                    <span className="score-label">Final Score:</span>
                    <span className="score-value">{s.overall_score || '--'}</span>
                  </div>
                  <GlowingButton 
                    variant="secondary" 
                    className="view-btn"
                    onClick={() => navigate(`/report/${s.session_id}`)}
                  >
                    <ExternalLink size={18} />
                  </GlowingButton>
                  <GlowingButton 
                    variant="danger" 
                    className="view-btn"
                    style={{ marginLeft: '12px' }}
                    onClick={() => setDeleteSessionTargetId(s.session_id)}
                    title="删除记录"
                  >
                    <Trash2 size={18} />
                  </GlowingButton>
                </div>
              ))}
            </div>
          )}
        </GlassCard>
      </div>
    </div>

    {/* --- Overlay Modals --- */}
      
      {/* 1. Resume Preview Modal */}
      {activePreviewResume && (
        <div className="modal-overlay" onClick={() => setActivePreviewResume(null)}>
          <div className="modal-content size-lg" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title">
                <FileText size={18} />
                <span>Resume Preview: {activePreviewResume.filename}</span>
              </h2>
              <button className="modal-close-btn" onClick={() => setActivePreviewResume(null)}>
                <X size={18} />
              </button>
            </div>
            <div className="modal-body">
              {activePreviewResume.downloadUrl ? (
                <div className="pdf-preview-container">
                  <iframe 
                    src={activePreviewResume.downloadUrl} 
                    className="pdf-iframe" 
                    title="Resume PDF Preview"
                  />
                </div>
              ) : (
                <div className="markdown-preview-scroll">
                  {activePreviewResume.rawText || "No content parsed."}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* 2. Job Creation Modal */}
      {showJobModal && (
        <div className="modal-overlay" onClick={() => setShowJobModal(false)}>
          <div className="modal-content size-md" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title">
                <Briefcase size={18} />
                <span>Configure Job Profile</span>
              </h2>
              <button className="modal-close-btn" onClick={() => setShowJobModal(false)}>
                <X size={18} />
              </button>
            </div>
            <div className="modal-body">
              <div className="modal-form">
                <div className="form-group">
                  <label htmlFor="job-title-input">Job Title</label>
                  <input 
                    id="job-title-input"
                    type="text" 
                    placeholder="e.g. Senior Frontend Engineer"
                    value={jobNameInput}
                    onChange={e => setJobNameInput(e.target.value)}
                    className="modern-input"
                  />
                </div>
                <div className="form-group">
                  <label htmlFor="job-desc-input">Job Description</label>
                  <textarea 
                    id="job-desc-input"
                    placeholder="Outline job requirements, tech stack, and responsibilities..."
                    value={jobDescInput}
                    onChange={e => setJobDescInput(e.target.value)}
                    className="modern-input"
                    rows={6}
                  />
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <GlowingButton 
                variant="secondary" 
                onClick={() => setShowJobModal(false)}
                disabled={creatingJob}
              >
                Cancel
              </GlowingButton>
              <GlowingButton 
                variant="primary" 
                onClick={handleCreateJob}
                disabled={!jobNameInput || !jobDescInput || creatingJob}
              >
                {creatingJob ? 'Configuring...' : 'Set Parameters'}
              </GlowingButton>
            </div>
          </div>
        </div>
      )}

      {/* 3. Job Preview Modal */}
      {activePreviewJob && (
        <div className="modal-overlay" onClick={() => setActivePreviewJob(null)}>
          <div className="modal-content size-md" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title">
                <Briefcase size={18} />
                <span>Job Details: {activePreviewJob.name}</span>
              </h2>
              <button className="modal-close-btn" onClick={() => setActivePreviewJob(null)}>
                <X size={18} />
              </button>
            </div>
            <div className="modal-body">
              <div className="markdown-preview-scroll">
                {activePreviewJob.description || "No description provided."}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 4. Start Session Selection Modal */}
      {showStartModal && (
        <div className="modal-overlay" onClick={() => setShowStartModal(false)}>
          <div className="modal-content size-md" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title">
                <PlayCircle size={18} />
                <span>Initialize Simulation Parameters</span>
              </h2>
              <button className="modal-close-btn" onClick={() => setShowStartModal(false)}>
                <X size={18} />
              </button>
            </div>
            <div className="modal-body">
              <div className="modal-form">
                <div className="form-group">
                  <label htmlFor="start-resume-select">Select Interviewee Resume</label>
                  {resumeHistory.length === 0 && !resumeId ? (
                    <div className="dashboard-error" style={{ margin: 0, padding: '12px', fontSize: '12px' }}>
                      No resumes found. Please upload a resume on the main dashboard first.
                    </div>
                  ) : (
                    <select
                      id="start-resume-select"
                      value={selectedStartResumeId}
                      onChange={e => setSelectedStartResumeId(e.target.value)}
                      className="modern-input"
                    >
                      <option value="">-- Choose a Resume --</option>
                      {resumeId && !resumeHistory.some(r => r.id === resumeId) && (
                        <option value={resumeId}>[Active] Current Uploaded Resume</option>
                      )}
                      {resumeHistory.map(r => (
                        <option key={r.id} value={r.id}>
                          {r.filename}
                        </option>
                      ))}
                    </select>
                  )}
                </div>

                <div className="form-group">
                  <label htmlFor="start-job-select">Select Target Job Profile</label>
                  {jobHistory.length === 0 && !jobId ? (
                    <div className="dashboard-error" style={{ margin: 0, padding: '12px', fontSize: '12px' }}>
                      No job profiles found. Please configure a job on the main dashboard first.
                    </div>
                  ) : (
                    <select
                      id="start-job-select"
                      value={selectedStartJobId}
                      onChange={e => setSelectedStartJobId(e.target.value)}
                      className="modern-input"
                    >
                      <option value="">-- Choose a Job Profile --</option>
                      {jobId && !jobHistory.some(j => j.id === jobId) && (
                        <option value={jobId}>[Active] {jobName}</option>
                      )}
                      {jobHistory.map(j => (
                        <option key={j.id} value={j.id}>
                          {j.name}
                        </option>
                      ))}
                    </select>
                  )}
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <GlowingButton 
                variant="secondary" 
                onClick={() => setShowStartModal(false)}
                disabled={starting}
              >
                Cancel
              </GlowingButton>
              <GlowingButton 
                variant="primary" 
                onClick={handleConfirmStartSession}
                disabled={!selectedStartResumeId || !selectedStartJobId || starting}
              >
                {starting ? 'Linking Neural Net...' : 'Confirm & Start Simulation'}
              </GlowingButton>
            </div>
          </div>
        </div>
      )}

      {/* 5. Delete Session Confirmation Modal */}
      {deleteSessionTargetId && (
        <div className="modal-overlay" onClick={() => setDeleteSessionTargetId('')}>
          <div className="modal-content size-sm" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title" style={{ color: 'var(--danger)' }}>
                <Trash2 size={18} style={{ color: 'var(--danger)' }} />
                <span>确认删除记录</span>
              </h2>
              <button className="modal-close-btn" onClick={() => setDeleteSessionTargetId('')}>
                <X size={18} />
              </button>
            </div>
            <div className="modal-body">
              <p style={{ color: 'var(--text-main)', fontSize: '1rem', lineHeight: '1.6', margin: 0 }}>
                确定要永久删除该面试记录吗？这将同时清空所有的答题音频与对话记录，此操作不可撤销。
              </p>
            </div>
            <div className="modal-footer">
              <GlowingButton 
                variant="secondary" 
                onClick={() => setDeleteSessionTargetId('')}
              >
                取消
              </GlowingButton>
              <GlowingButton 
                variant="danger" 
                onClick={handleConfirmDeleteSession}
              >
                确认删除
              </GlowingButton>
            </div>
          </div>
        </div>
      )}
    </>
  );
};
