class VideoUploader {
    constructor() {
        this.chunkSize = 1024 * 1024; // 1MB chunks
        this.uploadBtn = document.getElementById('uploadBtn');
        this.fileInput = document.getElementById('videoFile');
        this.progressBar = document.getElementById('uploadProgress');
        
        this.uploadBtn.addEventListener('click', () => this.startUpload());
    }
    
    async startUpload() {
        const file = this.fileInput.files[0];
        if (!file) {
            alert('Please select a file first');
            return;
        }
        
        try {
            // 初始化上传
            const response = await fetch('/api/upload/init', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    fileName: file.name,
                    fileSize: file.size,
                    contentType: file.type
                })
            });
            
            if (!response.ok) {
                throw new Error('Failed to initialize upload');
            }
            
            const { uploadId, chunkSize } = await response.json();
            
            // 分片上传
            const chunks = Math.ceil(file.size / chunkSize);
            for (let i = 0; i < chunks; i++) {
                const start = i * chunkSize;
                const end = Math.min(start + chunkSize, file.size);
                const chunk = file.slice(start, end);
                
                const formData = new FormData();
                formData.append('chunk', chunk);
                formData.append('uploadId', uploadId);
                formData.append('chunkIndex', i);
                
                const chunkResponse = await fetch('/api/upload/chunk', {
                    method: 'POST',
                    body: formData
                });
                
                if (!chunkResponse.ok) {
                    throw new Error(`Failed to upload chunk ${i}`);
                }
                
                // 更新进度条
                const progress = ((i + 1) / chunks) * 100;
                this.updateProgress(progress);
            }
            
            // 完成上传
            const completeResponse = await fetch('/api/upload/complete', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ uploadId })
            });
            
            if (!completeResponse.ok) {
                throw new Error('Failed to complete upload');
            }
            
            const result = await completeResponse.json();
            console.log('Upload completed:', result);
            
            // 等待视频处理完成
            await this.waitForVideoReady(uploadId);
            
            // 刷新视频列表
            if (window.videoList) {
                await window.videoList.loadVideos();
            } else {
                console.error('Video list not initialized');
            }
            
            alert('Upload completed successfully! Please restart the server to play the uploaded video.');
            
        } catch (error) {
            console.error('Upload failed:', error);
            alert('Upload failed: ' + error.message);
        }
    }
    
    async waitForVideoReady(videoId, maxAttempts = 30) {
        console.log('Waiting for video to be ready...');
        let attempts = 0;
        
        while (attempts < maxAttempts) {
            try {
                const response = await fetch(`/api/videos/${videoId}/info`);
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                
                const videoInfo = await response.json();
                console.log('Video status:', videoInfo.status);
                
                if (videoInfo.status === 'ready') {
                    console.log('Video is ready!');
                    return true;
                }
                
                if (videoInfo.status === 'error') {
                    throw new Error('Video processing failed');
                }
                
                // 等待 2 秒后重试
                await new Promise(resolve => setTimeout(resolve, 2000));
                attempts++;
                
            } catch (error) {
                console.error('Error checking video status:', error);
                throw error;
            }
        }
        
        throw new Error('Video processing timeout');
    }
    
    updateProgress(progress) {
        this.progressBar.style.width = `${progress}%`;
        
        // Update upload button state
        if (progress > 0 && progress < 100) {
            this.uploadBtn.disabled = true;
            this.uploadBtn.classList.add('opacity-50', 'cursor-not-allowed');
            this.uploadBtn.textContent = 'Uploading...';
        } else {
            this.uploadBtn.disabled = false;
            this.uploadBtn.classList.remove('opacity-50', 'cursor-not-allowed');
            this.uploadBtn.textContent = 'Upload';
        }
    }
}

// 初始化上传器
document.addEventListener('DOMContentLoaded', () => {
    window.videoUploader = new VideoUploader();
}); 