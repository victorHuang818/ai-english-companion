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
  RefreshCw,
  ArrowRight
} from 'lucide-react';
import './DashboardPage.css';
import '../components/DashboardModals.css';

interface ScenarioPreset {
  name: string;
  description: string;
  category: string;
  icon: React.ComponentType<any>;
  details: string;
}

const PRESET_SCENARIOS: ScenarioPreset[] = [
  {
    name: 'Job Interview',
    description: '外企求职面试',
    category: 'Career',
    icon: Briefcase,
    details: 'Simulate a professional interview at a multinational corporation. Focus on project experience, career goals, and behavioral questions.'
  },
  {
    name: 'Ordering Food',
    description: '西餐厅点餐',
    category: 'Daily Life',
    icon: Coffee,
    details: 'Practice ordering food, asking about the menu, and handling payments in a dining or restaurant context.'
  },
  {
    name: 'IELTS Speaking',
    description: '雅思口语模拟',
    category: 'Exam Prep',
    icon: BookOpen,
    details: 'Simulate IELTS speaking test sections (Part 1, Part 2, Part 3) under standardized constraints.'
  },
  {
    name: 'Business Meeting',
    description: '商务会议沟通',
    category: 'Workplace',
    icon: Users,
    details: 'Practice presenting an idea, reporting project status, or discussing proposals in a business meeting.'
  }
];

export const DashboardPage: React.FC = () => {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();

  // Profile configuration states
  const [englishLevel, setEnglishLevel] = useState<'beginner' | 'intermediate' | 'advanced'>('intermediate');
  const [learningTarget, setLearningTarget] = useState<string>('Improve spoken fluency and prepare for professional communication');
  const [savedProfileId, setSavedProfileId] = useState<string>('');
  const [savingProfile, setSavingProfile] = useState<boolean>(false);
  const [profileSavedSuccess, setProfileSavedSuccess] = useState<boolean>(false);

  // Scenario selection states
  const [selectedScenario, setSelectedScenario] = useState<ScenarioPreset | { name: string; description: string; isCustom: boolean } | null>(null);
  const [showCustomScenarioModal, setShowCustomScenarioModal] = useState<boolean>(false);
  const [customName, setCustomName] = useState<string>('');
  const [customDesc, setCustomDesc] = useState<string>('');
  const [customScenariosList, setCustomScenariosList] = useState<Array<{ id: string; name: string; description: string }>>([]);

  // Session states
  const [starting, setStarting] = useState(false);
  const [error, setError] = useState('');
  const [sessions, setSessions] = useState<any[]>([]);
  const [loadingSessions, setLoadingSessions] = useState(false);
  const [deleteSessionTargetId, setDeleteSessionTargetId] = useState<string>('');

  // Load profile & custom scenarios from DB / LocalStorage
  useEffect(() => {
    if (user?.id) {
      // Restore profile from local storage if exists
      const savedProfileKey = `user_profile_${user.id}`;
      const localProfile = localStorage.getItem(savedProfileKey);
      if (localProfile) {
        const parsed = JSON.parse(localProfile);
        setSavedProfileId(parsed.id);
        setEnglishLevel(parsed.english_level);
        setLearningTarget(parsed.learning_target);
        setProfileSavedSuccess(true);
      }

      // Restore custom scenarios list
      const savedScenariosKey = `custom_scenarios_${user.id}`;
      const localScenarios = localStorage.getItem(savedScenariosKey);
      if (localScenarios) {
        setCustomScenariosList(JSON.parse(localScenarios));
      }
    }
  }, [user]);

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
      
      if (user?.id) {
        localStorage.setItem(`user_profile_${user.id}`, JSON.stringify({
          id: res.id,
          english_level: englishLevel,
          learning_target: learningTarget
        }));
      }
    } catch (err: any) {
      setError('Failed to configure profile: ' + err.message);
    } finally {
      setSavingProfile(false);
    }
  };

  // Reset profile to allow editing
  const handleResetProfile = () => {
    setProfileSavedSuccess(false);
  };

  // Select a preset scenario card
  const handleSelectPreset = (preset: ScenarioPreset) => {
    setSelectedScenario(preset);
  };

  // Select a custom scenario
  const handleSelectCustomItem = (scenario: { id: string; name: string; description: string }) => {
    setSelectedScenario({
      name: scenario.name,
      description: scenario.description,
      isCustom: true
    });
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

      const newScenario = {
        id: res.id,
        name: customName,
        description: customDesc
      };

      const updatedList = [newScenario, ...customScenariosList];
      setCustomScenariosList(updatedList);
      if (user?.id) {
        localStorage.setItem(`custom_scenarios_${user.id}`, JSON.stringify(updatedList));
      }

      // Automatically select the newly created custom scenario
      setSelectedScenario({
        name: customName,
        description: customDesc,
        isCustom: true
      });
      
      setShowCustomScenarioModal(false);
      setCustomName('');
      setCustomDesc('');
    } catch (err: any) {
      setError('Failed to create scenario: ' + err.message);
    }
  };

  // Delete custom scenario from local cache list
  const handleDeleteCustomScenario = (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const filtered = customScenariosList.filter(item => item.id !== id);
    setCustomScenariosList(filtered);
    if (user?.id) {
      localStorage.setItem(`custom_scenarios_${user.id}`, JSON.stringify(filtered));
    }
    setSelectedScenario(null);
  };

  // Start Session (Create Profile/Scenario if not yet done, then create session)
  const handleStartPractice = async () => {
    if (!profileSavedSuccess || !englishLevel || !learningTarget) {
      setError('Please configure your English Profile first.');
      return;
    }
    if (!selectedScenario) {
      setError('Please choose a practice scenario.');
      return;
    }

    try {
      setStarting(true);
      setError('');

      // 1. Ensure user profile is saved and ID is active
      let profileId = savedProfileId;
      if (!profileId) {
        const resProfile = await api.createUserProfile({
          english_level: englishLevel,
          learning_target: learningTarget
        });
        profileId = resProfile.id;
        setSavedProfileId(profileId);
        if (user?.id) {
          localStorage.setItem(`user_profile_${user.id}`, JSON.stringify({
            id: profileId,
            english_level: englishLevel,
            learning_target: learningTarget
          }));
        }
      }

      // 2. Resolve or create scenario ID
      let scenarioId = '';
      const scenarioName = selectedScenario.name;
      const scenarioDesc = 'details' in selectedScenario ? selectedScenario.details : selectedScenario.description;

      // Check if this scenario name is already created in our local list of scenarios
      const cachedKey = `scenario_id_${user?.id}_${scenarioName}`;
      const cachedScenarioId = localStorage.getItem(cachedKey);
      
      if (cachedScenarioId) {
        scenarioId = cachedScenarioId;
      } else {
        // Find if it is in custom scenario list
        const customMatch = customScenariosList.find(s => s.name === scenarioName);
        if (customMatch) {
          scenarioId = customMatch.id;
        } else {
          // Create scenario in DB
          const resScenario = await api.createScenario({
            name: scenarioName,
            description: scenarioDesc
          });
          scenarioId = resScenario.id;
          localStorage.setItem(cachedKey, scenarioId);
        }
      }

      // 3. Create practice session
      const resSession = await api.createSession({
        user_profile_id: profileId,
        scenario_id: scenarioId
      });

      navigate(`/companion/${resSession.session_id}`);
    } catch (err: any) {
      setError('Failed to launch practice room: ' + err.message);
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
              {profileSavedSuccess && (
                <button className="reset-btn" onClick={handleResetProfile} title="Edit Profile">
                  <RefreshCw size={14} />
                </button>
              )}
            </div>
            <div className="card-content">
              {profileSavedSuccess ? (
                <div className="profile-saved-card animate-fade-in">
                  <div className="profile-status">
                    <CheckCircle className="check-icon" size={18} />
                    <span>Active Profile Configured</span>
                  </div>
                  <div className="profile-detail-item">
                    <span className="detail-label">Current English Level:</span>
                    <span className="detail-value level-badge capitalize">{englishLevel}</span>
                  </div>
                  <div className="profile-detail-item">
                    <span className="detail-label">Learning Goal / Focus:</span>
                    <p className="detail-text">{learningTarget}</p>
                  </div>
                </div>
              ) : (
                <div className="form-section">
                  <div className="form-group">
                    <label>Select English Level</label>
                    <div className="level-select-row">
                      {(['beginner', 'intermediate', 'advanced'] as const).map(level => (
                        <button
                          key={level}
                          type="button"
                          className={`level-btn capitalize ${englishLevel === level ? 'active' : ''}`}
                          onClick={() => setEnglishLevel(level)}
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
                      onChange={e => setLearningTarget(e.target.value)}
                      rows={4}
                    />
                  </div>
                  <GlowingButton
                    variant="secondary"
                    onClick={handleSaveProfile}
                    disabled={savingProfile || !learningTarget.trim()}
                    fullWidth
                  >
                    {savingProfile ? 'Saving profile...' : 'Save Profile'}
                  </GlowingButton>
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
                <button className="reset-btn" onClick={() => setSelectedScenario(null)} title="Clear Selection">
                  <RefreshCw size={14} />
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
                    <p className="details">
                      {'details' in selectedScenario ? selectedScenario.details : selectedScenario.description}
                    </p>
                  </div>
                </div>
              ) : (
                <div className="scenarios-selection-flow">
                  <div className="scenarios-grid-mini">
                    {PRESET_SCENARIOS.map(preset => {
                      const Icon = preset.icon;
                      return (
                        <div
                          key={preset.name}
                          className="scenario-preset-card"
                          onClick={() => handleSelectPreset(preset)}
                        >
                          <div className="preset-icon-container">
                            <Icon size={20} />
                          </div>
                          <div className="preset-info">
                            <span className="preset-category">{preset.category}</span>
                            <span className="preset-name">{preset.name}</span>
                            <span className="preset-desc">{preset.description}</span>
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
                            >
                              <Compass size={16} className="item-icon" />
                              <div className="job-info-block">
                                <span className="item-filename">{item.name}</span>
                                <span className="item-sub-desc">{item.description}</span>
                              </div>
                              <div className="item-actions">
                                <button
                                  className="action-btn delete"
                                  title="Delete Scenario"
                                  onClick={(e) => handleDeleteCustomScenario(item.id, e)}
                                >
                                  <Trash2 size={14} />
                                </button>
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
            className={`dashboard-card action-card ${profileSavedSuccess && selectedScenario ? 'ready' : ''}`} 
            glow={profileSavedSuccess && !!selectedScenario}
          >
            <div className="card-header">
              <Sparkles className="card-icon" />
              <h2>Step 3: Begin Companion Session</h2>
            </div>
            <div className="card-content centered">
              <p className="status-text">
                {profileSavedSuccess && selectedScenario
                  ? 'All parameters established. Companion interface ready.' 
                  : 'Establish background profile and choose scenario to initiate companion...'}
              </p>
              <GlowingButton 
                variant="primary" 
                onClick={handleStartPractice}
                disabled={starting || !profileSavedSuccess || !selectedScenario}
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
                          <Calendar size={14} /> {new Date(s.created_at * 1000).toLocaleDateString()}
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
    </>
  );
};
