import { useState, useRef } from 'react';
import { useNickname } from '../hooks/useNickname';
import { createTopic } from '../services/api';

interface CreateTopicFormProps {
  onSuccess?: () => void;
}

export function CreateTopicForm({ onSuccess }: CreateTopicFormProps) {
  const { displayName } = useNickname();
  const [title, setTitle] = useState('');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isDragOver, setIsDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileSelect = (file: File) => {
    if (!file.type.startsWith('image/')) {
      setError('Please select an image file');
      return;
    }

    if (file.size > 10 * 1024 * 1024) {
      setError('File size must be less than 10MB');
      return;
    }

    setSelectedFile(file);
    setError(null);

    const reader = new FileReader();
    reader.onloadend = () => {
      setPreview(reader.result as string);
    };
    reader.readAsDataURL(file);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
    
    const file = e.dataTransfer.files[0];
    if (file) {
      handleFileSelect(file);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!title.trim()) {
      setError('Please enter a title');
      return;
    }

    if (!selectedFile) {
      setError('Please select an image');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      await createTopic(title.trim(), selectedFile, displayName);
      setTitle('');
      setSelectedFile(null);
      setPreview(null);
      onSuccess?.();
    } catch {
      setError('Failed to create topic. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const clearSelection = () => {
    setSelectedFile(null);
    setPreview(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  return (
    <form onSubmit={handleSubmit} className="create-topic-form">
      <div 
        className={`drop-zone ${isDragOver ? 'drop-zone--active' : ''} ${preview ? 'drop-zone--has-preview' : ''}`}
        onDragOver={(e) => { e.preventDefault(); setIsDragOver(true); }}
        onDragLeave={() => setIsDragOver(false)}
        onDrop={handleDrop}
        onClick={() => fileInputRef.current?.click()}
      >
        {preview ? (
          <div className="drop-zone__preview">
            <img src={preview} alt="Preview" />
            <button 
              type="button" 
              className="drop-zone__clear"
              onClick={(e) => { e.stopPropagation(); clearSelection(); }}
            >
              ✕
            </button>
          </div>
        ) : (
          <div className="drop-zone__placeholder">
            <span className="drop-zone__icon">⬆</span>
            <span className="drop-zone__text">Drop image here or click to upload</span>
            <span className="drop-zone__hint">JPG, PNG, GIF, WEBP up to 10MB</span>
          </div>
        )}
        <input
          ref={fileInputRef}
          type="file"
          accept="image/jpeg,image/png,image/gif,image/webp"
          onChange={(e) => e.target.files?.[0] && handleFileSelect(e.target.files[0])}
          hidden
        />
      </div>

      <div className="form-group">
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Enter a title for your topic..."
          className="input-title"
          maxLength={200}
        />
      </div>

      {error && <div className="form-error">{error}</div>}

      <button 
        type="submit" 
        className="btn-submit"
        disabled={isSubmitting || !title.trim() || !selectedFile}
      >
        {isSubmitting ? 'Posting...' : 'Post Topic'}
      </button>
    </form>
  );
}

