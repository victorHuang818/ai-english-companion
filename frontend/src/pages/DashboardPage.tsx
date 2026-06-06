import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { GlassCard } from '../components/GlassCard';
import { GlowingButton } from '../components/GlowingButton';
import { api } from '../services/api';
import { useAuthStore } from '../store/authStore';
import { 
  LogOut, 
  History, 
  ExternalLink, 
  Calendar, 
  Trash2, 
  Plus, 
  X,
  CheckCircle,
  Award,
  Briefcase,
  Coffee,
  BookOpen,
  Users,
  Compass,
  Sparkles,
  ArrowRight,
  ArrowLeft
} from 'lucide-react';
import './DashboardPage.css';
import '../components/DashboardModals.css';

const SCENARIO_METADATA_MAP: Record<string, { category: string; subDesc: string; icon: React.ComponentType<any> }> = {
  'Job Interview': { category: 'Career', subDesc: '外企求职面试', icon: Briefcase },
  'Ordering Food': { category: 'Daily Life', subDesc: '西餐厅点餐', icon: Coffee },
  'IELTS Speaking': { category: 'Exam Prep', subDesc: '雅思口语模拟', icon: BookOpen },
  'Business Meeting': { category: 'Workplace', subDesc: '商务会议沟通', icon: Users }
};

export const DashboardPage: React.FC = () => {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();

  // Profile configuration states
  const [englishLevel, setEnglishLevel] = useState<'beginner' | 'intermediate' | 'advanced'>('intermediate');
  const [learningTarget, setLearningTarget] = useState<string>('Improve spoken fluency and prepare for professional communication');
  const [savedProfileId, setSavedProfileId] = useState<string>('');
  const [savingProfile, setSavingProfile] = useState<boolean>(false);
  const [profileSavedSuccess, setProfileSavedSuccess] = useState<boolean>(false);
  const [savedProfilesList, setSavedProfilesList] = useState<Array<{ id: string; english_level: 'beginner' | 'intermediate' | 'advanced'; learning_target: string }>>([]);

  // Scenario selection states
  const [scenariosList, setScenariosList] = useState<any[]>([]);
  const [selectedScenario, setSelectedScenario] = useState<any | null>(null);
  const [showCustomScenarioModal, setShowCustomScenarioModal] = useState<boolean>(false);
  const [customName, setCustomName] = useState<string>('');
  const [customDesc, setCustomDesc] = useState<string>('');

  // Start practice modal states
  const [showStartPracticeModal, setShowStartPracticeModal] = useState<boolean>(false);
  const [modalProfileId, setModalProfileId] = useState<string>('');
  const [modalScenarioId, setModalScenarioId] = useState<string>('');

  // Derived scenario lists
  const presetScenarios = scenariosList.filter(
    (s) => s.creator_id === '00000000-0000-0000-0000-000000000000'
  );
  const customScenariosList = scenariosList.filter(
    (s) => s.creator_id !== '00000000-0000-0000-0000-000000000000'
  );

  // Session states
  const [starting, setStarting] = useState(false);
  const [error, setError] = useState('');
  const [sessions, setSessions] = useState<any[]>([]);
  const [loadingSessions, setLoadingSessions] = useState(false);
  const [deleteSessionTargetId, setDeleteSessionTargetId] = useState<string>('');

  // Load profiles & scenarios from DB on mount
  useEffect(() => {
    if (user?.id) {
      fetchProfiles();
      fetchScenarios();
    }
  }, [user]);

  const fetchProfiles = async () => {
    try {
      const res = await api.listUserProfiles();
      const list = res.profiles || [];
      setSavedProfilesList(list);

      // Auto-load the most recent profile as the active one
      if (list.length > 0) {
        const parsed = list[0];
        setSavedProfileId(parsed.id);
        setEnglishLevel(parsed.english_level || 'intermediate');
        setLearningTarget(parsed.learning_target || '');
        setProfileSavedSuccess(true);
      } else {
        setSavedProfileId('');
        setEnglishLevel('intermediate');
        setLearningTarget('Improve spoken fluency and prepare for professional communication');
        setProfileSavedSuccess(false);
      }
    } catch (err) {
      console.error('Failed to fetch user profiles:', err);
    }
  };

  const fetchScenarios = async () => {
    try {
      const res = await api.listScenarios();
      setScenariosList(res.scenarios || []);
    } catch (err) {
      console.error('Failed to fetch scenarios:', err);
    }
  };

  // Track profile changes to prompt for re-saving
  const handleLevelChange = (level: 'beginner' | 'intermediate' | 'advanced') => {
    setEnglishLevel(level);
    setProfileSavedSuccess(false);
  };

  const handleTargetChange = (target: string) => {
    setLearningTarget(target);
    setProfileSavedSuccess(false);
  };

  const handleSelectProfile = (profile: { id: string; english_level: 'beginner' | 'intermediate' | 'advanced'; learning_target: string }) => {
    setSavedProfileId(profile.id);
    setEnglishLevel(profile.english_level);
    setLearningTarget(profile.learning_target);
    setProfileSavedSuccess(true);
  };

  // Save/Update English Profile
  const handleSaveProfile = async () => {
    try {
      setSavingProfile(true);
      setError('');
      const res = await api.createUserProfile({
        english_level: englishLevel,
        learning_target: learningTarget
      });
      
      setSavedProfileId(res.id);
      setProfileSavedSuccess(true);
      
      // Refresh the list of saved profiles
      await fetchProfiles();
    } catch (err: any) {
      setError('Failed to configure profile: ' + err.message);
    } finally {
      setSavingProfile(false);
    }
  };

  // Load practice sessions
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

  // Select a preset scenario card
  const handleSelectPreset = (preset: any) => {
    setSelectedScenario(preset);
  };

  // Select a custom scenario
  const handleSelectCustomItem = (scenario: any) => {
    setSelectedScenario(scenario);
  };

  // Create custom scenario
  const handleCreateCustomScenario = async () => {
    if (!customName || !customDesc) return;
    try {
      setError('');
      const res = await api.createScenario({
        name: customName,
        description: customDesc
      });

      // Refresh scenarios list
      await fetchScenarios();

      // Automatically select the newly created custom scenario
      setSelectedScenario({
        id: res.id,
        name: customName,
        description: customDesc,
        creator_id: user?.id || ''
      });
      
      setShowCustomScenarioModal(false);
      setCustomName('');
      setCustomDesc('');
    } catch (err: any) {
      setError('Failed to create scenario: ' + err.message);
    }
  };

  // Open Start Modal
  const handleOpenStartModal = () => {
    const initialProfileId = savedProfileId || (savedProfilesList.length > 0 ? savedProfilesList[0].id : '');
    const initialScenarioId = selectedScenario?.id || (scenariosList.length > 0 ? scenariosList[0].id : '');
    
    setModalProfileId(initialProfileId);
    setModalScenarioId(initialScenarioId);
    setShowStartPracticeModal(true);
  };

  // Start Session from the Confirmation Modal
  const handleStartPracticeFromModal = async () => {
    if (!modalProfileId || !modalScenarioId) return;
    try {
      setStarting(true);
      setError('');

      // Create practice session
      const resSession = await api.createSession({
        user_profile_id: modalProfileId,
        scenario_id: modalScenarioId
      });

      // Close modal
      setShowStartPracticeModal(false);

      navigate(`/companion/${resSession.session_id}`);
    } catch (err: any) {
      setError('Failed to launch practice room: ' + err.message);
      setShowStartPracticeModal(false);
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
      setError('Failed to delete practice session: ' + err.message);
      setDeleteSessionTargetId('');
    }
  };

  return (
    <>
      <div className="dashboard-container animate-fade-in">
        <nav className="dashboard-nav glass-panel">
          <div className="nav-brand">
            <div className="brand-logo">EN</div>
            <span>AI English Companion</span>
          </div>
          <div className="nav-user">
            <div className="token-badge">
              <span className="token-label">Tokens Left:</span>
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
          {/* Step 1: English Profile */}
          <GlassCard className="dashboard-card" glow={!profileSavedSuccess}>
            <div className="card-header">
              <Award className="card-icon" />
              <h2>Step 1: Background English Profile</h2>
            </div>
            <div className="card-content">
              <div className="form-section">
                <div className="form-group">
                  <label>Select English Level</label>
                  <div className="level-select-row">
                    {(['beginner', 'intermediate', 'advanced'] as const).map(level => (
                      <button
                        key={level}
                        type="button"
                        className={`level-btn capitalize ${englishLevel === level ? 'active' : ''}`}
                        onClick={() => handleLevelChange(level)}
                      >
                        {level === 'beginner' && '🌱 '}
                        {level === 'intermediate' && '🚀 '}
                        {level === 'advanced' && '🏆 '}
                        {level}
                      </button>
                    ))}
                  </div>
                </div>
                <div className="form-group">
                  <label htmlFor="learning-target">Practice Target or context</label>
                  <textarea
                    id="learning-target"
                    className="modern-input"
                    placeholder="e.g., I want to prepare for an upcoming overseas trip, focus on accent, vocabulary, and basic hotel check-in / shopping scenarios."
                    value={learningTarget}
                    onChange={e => handleTargetChange(e.target.value)}
                    rows={4}
                  />
                </div>
                <GlowingButton
                  variant="primary"
                  onClick={handleSaveProfile}
                  disabled={savingProfile || !learningTarget.trim()}
                  fullWidth
                >
                  {savingProfile ? 'Saving profile...' : 'Save Profile'}
                </GlowingButton>
              </div>

              {/* Previously Saved Profiles */}
              {savedProfilesList.length > 0 && (
                <div className="custom-scenarios-history" style={{ marginTop: '20px', borderTop: '1px solid var(--border-glass)', paddingTop: '16px' }}>
                  <div className="history-list-title">PREVIOUSLY SAVED PROFILES</div>
                  <div className="history-items-scroll">
                    {savedProfilesList.map(profile => {
                      const abbreviatedTarget = profile.learning_target.length > 25 
                        ? profile.learning_target.substring(0, 25) + '...' 
                        : profile.learning_target;
                      const displayTitle = `${profile.english_level} - ${abbreviatedTarget}`;
                      
                      return (
                        <div
                          key={profile.id}
                          className={`history-item ${savedProfileId === profile.id && profileSavedSuccess ? 'active-profile-item' : ''}`}
                          onClick={() => handleSelectProfile(profile)}
                          style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}
                        >
                          <Award size={16} className="item-icon" />
                          <span className="item-filename" style={{ fontSize: '13px' }}>{displayTitle}</span>
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}
            </div>
          </GlassCard>

          {/* Step 2: Practice Scenario */}
          <GlassCard className="dashboard-card" glow={profileSavedSuccess && !selectedScenario}>
            <div className="card-header">
              <Compass className="card-icon" />
              <h2>Step 2: Choose Speaking Scenario</h2>
              {selectedScenario && (
                <button 
                  className="back-btn" 
                  onClick={() => setSelectedScenario(null)} 
                  title="Back to Scenarios"
                  style={{
                    background: 'none',
                    border: 'none',
                    padding: 0,
                    margin: 0,
                    marginLeft: 'auto',
                    cursor: 'pointer',
                    display: 'inline-flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    boxShadow: 'none',
                    borderRadius: 0,
                    outline: 'none'
                  }}
                >
                  <ArrowLeft size={16} style={{ color: '#000000' }} />
                </button>
              )}
            </div>
            <div className="card-content">
              {selectedScenario ? (
                <div className="scenario-selected-preview animate-fade-in">
                  <div className="scenario-selected-header">
                    <CheckCircle className="check-icon" size={18} />
                    <span>Scenario Selected</span>
                  </div>
                  <div className="scenario-selected-body">
                    <h3>{selectedScenario.name}</h3>
                    <p className="description">{selectedScenario.description}</p>
                  </div>
                </div>
              ) : (
                <div className="scenarios-selection-flow">
                  <div className="scenarios-grid-mini">
                    {presetScenarios.map(preset => {
                      const meta = SCENARIO_METADATA_MAP[preset.name] || {
                        category: 'General',
                        subDesc: '日常口语练习',
                        icon: Compass
                      };
                      const Icon = meta.icon;
                      return (
                        <div
                          key={preset.id}
                          className="scenario-preset-card"
                          onClick={() => handleSelectPreset(preset)}
                        >
                          <div className="preset-icon-container">
                            <Icon size={20} />
                          </div>
                          <div className="preset-info">
                            <span className="preset-category">{meta.category}</span>
                            <span className="preset-name">{preset.name}</span>
                            <span className="preset-desc">{meta.subDesc}</span>
                          </div>
                        </div>
                      );
                    })}
                  </div>

                  <div className="custom-scenarios-section">
                    <button
                      className="add-custom-scenario-btn"
                      onClick={() => setShowCustomScenarioModal(true)}
                    >
                      <Plus size={16} />
                      <span>Configure Custom Scenario</span>
                    </button>

                    {customScenariosList.length > 0 && (
                      <div className="custom-scenarios-history">
                        <div className="history-list-title">Your Custom Scenarios</div>
                        <div className="history-items-scroll">
                          {customScenariosList.map(item => (
                            <div
                              key={item.id}
                              className="history-item"
                              onClick={() => handleSelectCustomItem(item)}
                              style={{ cursor: 'pointer' }}
                            >
                              <Compass size={16} className="item-icon" />
                              <div className="job-info-block">
                                <span className="item-filename">{item.name}</span>
                                <span className="item-sub-desc">{item.description}</span>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          </GlassCard>

          {/* Step 3: Start Practice */}
          <GlassCard 
            className="dashboard-card action-card ready" 
            glow={true}
          >
            <div className="card-header">
              <Sparkles className="card-icon" />
              <h2>Step 3: Begin Companion Session</h2>
            </div>
            <div className="card-content centered">
              <p className="status-text">
                {profileSavedSuccess && selectedScenario
                  ? `Selected: [${englishLevel.toUpperCase()}] and [${selectedScenario.name}]. Click below to confirm and start.` 
                  : 'Click below to select your profile and scenario and initiate the companion session.'}
              </p>
              <GlowingButton 
                variant="primary" 
                onClick={handleOpenStartModal}
                disabled={starting}
                className="start-btn"
              >
                {starting ? 'Connecting AI...' : (
                  <span className="start-btn-content">
                    Start Speaking Practice <ArrowRight size={18} style={{ marginLeft: '8px' }} />
                  </span>
                )}
              </GlowingButton>
            </div>
          </GlassCard>
        </div>

        {/* Practice Session History */}
        <div className="history-section">
          <div className="section-header">
            <History size={24} className="accent-icon" />
            <h2>Spoken English Practice History</h2>
          </div>
          
          <GlassCard className="history-list-card">
            {loadingSessions ? (
              <div className="loading-state">Loading history...</div>
            ) : sessions.length === 0 ? (
              <div className="empty-state">No past companion sessions found. Start your first session above!</div>
            ) : (
              <div className="sessions-list">
                {sessions.map(s => (
                  <div key={s.session_id} className="session-row">
                    <div className="session-info">
                      <span className="session-job">{s.scenario_name || 'Speaking Practice'}</span>
                      <div className="session-meta">
                        <span className="meta-item">
                          <Calendar size={14} /> {(() => {
                            const date = new Date(s.created_at * 1000);
                            const yyyy = date.getFullYear();
                            const mm = String(date.getMonth() + 1).padStart(2, '0');
                            const dd = String(date.getDate()).padStart(2, '0');
                            const hh = String(date.getHours()).padStart(2, '0');
                            const min = String(date.getMinutes()).padStart(2, '0');
                            return `${yyyy}-${mm}-${dd} ${hh}:${min}`;
                          })()}
                        </span>
                        <span className={`status-tag ${s.status.toLowerCase()}`}>{s.status}</span>
                      </div>
                    </div>
                    <div className="session-score">
                      <span className="score-label">Overall Score:</span>
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
                      title="Delete practice record"
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

      {/* --- Custom Scenario Modal --- */}
      {showCustomScenarioModal && (
        <div className="modal-overlay" onClick={() => setShowCustomScenarioModal(false)}>
          <div className="modal-content size-md" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title">
                <Compass size={18} />
                <span>Configure Custom Scenario</span>
              </h2>
              <button className="modal-close-btn" onClick={() => setShowCustomScenarioModal(false)}>
                <X size={18} />
              </button>
            </div>
            <div className="modal-body">
              <div className="modal-form">
                <div className="form-group">
                  <label htmlFor="custom-scenario-name">Scenario Name</label>
                  <input 
                    id="custom-scenario-name"
                    type="text" 
                    placeholder="e.g. Hotel Check-in"
                    value={customName}
                    onChange={e => setCustomName(e.target.value)}
                    className="modern-input"
                  />
                </div>
                <div className="form-group">
                  <label htmlFor="custom-scenario-desc">Context & Role Guide (Prompt Description)</label>
                  <textarea 
                    id="custom-scenario-desc"
                    placeholder="Describe the context, who the AI role is, what task the user should complete..."
                    value={customDesc}
                    onChange={e => setCustomDesc(e.target.value)}
                    className="modern-input"
                    rows={6}
                  />
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <GlowingButton 
                variant="secondary" 
                onClick={() => setShowCustomScenarioModal(false)}
              >
                Cancel
              </GlowingButton>
              <GlowingButton 
                variant="primary" 
                onClick={handleCreateCustomScenario}
                disabled={!customName.trim() || !customDesc.trim()}
              >
                Create Scenario
              </GlowingButton>
            </div>
          </div>
        </div>
      )}

      {/* --- Delete Session Confirmation Modal --- */}
      {deleteSessionTargetId && (
        <div className="modal-overlay" onClick={() => setDeleteSessionTargetId('')}>
          <div className="modal-content size-sm" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title" style={{ color: 'var(--danger)' }}>
                <Trash2 size={18} style={{ color: 'var(--danger)' }} />
                <span>Delete Practice Session</span>
              </h2>
              <button className="modal-close-btn" onClick={() => setDeleteSessionTargetId('')}>
                <X size={18} />
              </button>
            </div>
            <div className="modal-body">
              <p style={{ color: 'var(--text-main)', fontSize: '1rem', lineHeight: '1.6', margin: 0 }}>
                Are you sure you want to permanently delete this speaking practice record? All audio data and conversation history will be lost. This action is irreversible.
              </p>
            </div>
            <div className="modal-footer">
              <GlowingButton 
                variant="secondary" 
                onClick={() => setDeleteSessionTargetId('')}
              >
                Cancel
              </GlowingButton>
              <GlowingButton 
                variant="danger" 
                onClick={handleConfirmDeleteSession}
              >
                Confirm Delete
              </GlowingButton>
            </div>
          </div>
        </div>
      )}

      {/* --- Start Practice Modal --- */}
      {showStartPracticeModal && (
        <div className="modal-overlay" onClick={() => setShowStartPracticeModal(false)}>
          <div className="modal-content size-md" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title">
                <Sparkles size={18} />
                <span>Start Speaking Practice</span>
              </h2>
              <button className="modal-close-btn" onClick={() => setShowStartPracticeModal(false)}>
                <X size={18} />
              </button>
            </div>
            <div className="modal-body">
              <div className="modal-form">
                <div className="form-group">
                  <label htmlFor="modal-profile-select">Select English Profile</label>
                  {savedProfilesList.length === 0 ? (
                    <div style={{ color: 'var(--danger)', fontSize: '0.9rem', marginTop: '4px' }}>
                      ⚠️ No saved profiles found. Please configure and save a profile in Step 1 first.
                    </div>
                  ) : (
                    <>
                      <select 
                        id="modal-profile-select"
                        value={modalProfileId}
                        onChange={e => setModalProfileId(e.target.value)}
                        className="modern-input"
                        style={{ width: '100%', marginBottom: '8px' }}
                      >
                        {savedProfilesList.map(p => (
                          <option key={p.id} value={p.id}>
                            {p.english_level.toUpperCase()} - {p.learning_target.substring(0, 30)}{p.learning_target.length > 30 ? '...' : ''}
                          </option>
                        ))}
                      </select>
                      {(() => {
                        const selectedP = savedProfilesList.find(p => p.id === modalProfileId);
                        if (!selectedP) return null;
                        return (
                          <div style={{ 
                            background: 'rgba(0, 102, 204, 0.04)', 
                            border: '1px solid var(--border-glass)', 
                            padding: '10px', 
                            borderRadius: '8px',
                            fontSize: '0.85rem',
                            color: 'var(--text-muted)'
                          }}>
                            <strong>Level:</strong> <span className="capitalize">{selectedP.english_level}</span><br />
                            <strong>Target:</strong> {selectedP.learning_target}
                          </div>
                        );
                      })()}
                    </>
                  )}
                </div>

                <div className="form-group" style={{ marginTop: '16px' }}>
                  <label htmlFor="modal-scenario-select">Select Practice Scenario</label>
                  {scenariosList.length === 0 ? (
                    <div style={{ color: 'var(--danger)', fontSize: '0.9rem', marginTop: '4px' }}>
                      ⚠️ No scenarios found.
                    </div>
                  ) : (
                    <>
                      <select 
                        id="modal-scenario-select"
                        value={modalScenarioId}
                        onChange={e => setModalScenarioId(e.target.value)}
                        className="modern-input"
                        style={{ width: '100%', marginBottom: '8px' }}
                      >
                        {presetScenarios.length > 0 && (
                          <optgroup label="Preset Scenarios">
                            {presetScenarios.map(s => (
                              <option key={s.id} value={s.id}>{s.name}</option>
                            ))}
                          </optgroup>
                        )}
                        {customScenariosList.length > 0 && (
                          <optgroup label="Custom Scenarios">
                            {customScenariosList.map(s => (
                              <option key={s.id} value={s.id}>{s.name}</option>
                            ))}
                          </optgroup>
                        )}
                      </select>
                      {(() => {
                        const selectedS = scenariosList.find(s => s.id === modalScenarioId);
                        if (!selectedS) return null;
                        return (
                          <div style={{ 
                            background: 'rgba(92, 45, 145, 0.04)', 
                            border: '1px solid var(--border-glass)', 
                            padding: '10px', 
                            borderRadius: '8px',
                            fontSize: '0.85rem',
                            color: 'var(--text-muted)'
                          }}>
                            <strong>Scenario Name:</strong> {selectedS.name}<br />
                            <strong>Description:</strong> {selectedS.description}
                          </div>
                        );
                      })()}
                    </>
                  )}
                </div>
              </div>
            </div>
            <div className="modal-footer">
              <GlowingButton 
                variant="secondary" 
                onClick={() => setShowStartPracticeModal(false)}
              >
                Cancel
              </GlowingButton>
              <GlowingButton 
                variant="primary" 
                onClick={handleStartPracticeFromModal}
                disabled={starting || !modalProfileId || !modalScenarioId}
              >
                {starting ? 'Connecting AI...' : 'Confirm & Start'}
              </GlowingButton>
            </div>
          </div>
        </div>
      )}
    </>
  );
};
