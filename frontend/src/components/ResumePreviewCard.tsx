import React from 'react';
import { User, ShieldCheck, Cpu, GraduationCap, Eye } from 'lucide-react';
import './ResumePreviewCard.css';

interface ResumeData {
  name?: string;
  summary?: string;
  skills?: string[];
  experience_years?: number;
  education?: string;
}

interface ResumePreviewCardProps {
  data: string; // JSON string or raw markdown
  downloadUrl?: string;
  onPreviewClick: () => void;
}

export const ResumePreviewCard: React.FC<ResumePreviewCardProps> = ({ 
  data, 
  onPreviewClick 
}) => {
  let parsed: ResumeData | null = null;
  let isJson = false;

  if (data) {
    try {
      // Clean up string check - check if it looks like a JSON object
      if (data.trim().startsWith('{')) {
        parsed = JSON.parse(data);
        isJson = true;
      }
    } catch (e) {
      console.warn("Failed to parse resume content as JSON, falling back to markdown view.", e);
    }
  }

  return (
    <div className="resume-preview-container animate-fade-in">
      <div className="preview-header">
        <ShieldCheck className="accent-icon" />
        <h3>Neural Identity Profile</h3>
        <button className="preview-action-btn" onClick={onPreviewClick} title="Preview Resume Document">
          <Eye size={16} />
          <span>Preview PDF</span>
        </button>
      </div>
      
      {isJson && parsed ? (
        <div className="preview-grid">
          <div className="preview-item">
            <div className="item-label"><User size={14} /> Subject Name</div>
            <div className="item-value">{parsed.name || 'Unknown Entity'}</div>
          </div>
          
          <div className="preview-item">
            <div className="item-label"><Cpu size={14} /> Core Competencies</div>
            <div className="skills-tags">
              {parsed.skills?.map((s, i) => (
                <span key={i} className="skill-tag">{s}</span>
              )) || <span className="muted">None detected</span>}
            </div>
          </div>

          <div className="preview-item">
            <div className="item-label"><GraduationCap size={14} /> Education & Seniority</div>
            <div className="item-value">
              {parsed.experience_years ? `${parsed.experience_years} Years EXP` : 'N/A'} 
              {parsed.education && ` | ${parsed.education}`}
            </div>
          </div>
        </div>
      ) : (
        <div className="preview-markdown-fallback">
          <div className="fallback-badge">RAW MARKDOWN DETECTED</div>
          <p className="fallback-desc">
            The resume has been successfully parsed into Markdown format. Click the preview button above to view the full document.
          </p>
          <div className="markdown-snippet">
            {data ? data.slice(0, 180) + (data.length > 180 ? '...' : '') : 'Empty content'}
          </div>
        </div>
      )}

      {isJson && parsed?.summary && (
        <div className="preview-summary">
          <p>{parsed.summary}</p>
        </div>
      )}
    </div>
  );
};

