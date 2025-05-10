class VideoPlayer {
    constructor() {
        this.player = document.getElementById('videoPlayer');
        this.setupPlayer();
    }
    
    setupPlayer() {
        // 添加自定义控制
        this.player.addEventListener('loadedmetadata', () => {
            console.log('Video metadata loaded');
        });
        
        this.player.addEventListener('error', (e) => {
            console.error('Video error:', e);
            console.error('Error code:', this.player.error.code);
            console.error('Error message:', this.player.error.message);
            
            // 添加重试逻辑
            if (this.player.error.code === 4) { // DEMUXER_ERROR
                console.log('Attempting to reload video...');
                this.retryLoad();
            }
        });

        // 添加加载事件监听
        this.player.addEventListener('canplay', () => {
            console.log('Video can play');
            this.player.play().catch(error => {
                console.error('Play failed:', error);
            });
        });

        // 添加质量切换功能
        this.setupQualitySelector = () => {
            // 这个方法会在获取到视频信息后被调用
            console.log('Quality selector will be set up when video info is loaded');
        };
    }
    
    async loadVideo(videoId) {
        try {
            console.log('Loading video:', videoId);
            
            // 先检查视频状态
            const statusResponse = await fetch(`/api/videos/${videoId}/info`);
            if (!statusResponse.ok) {
                throw new Error(`HTTP error! status: ${statusResponse.status}`);
            }
            
            const videoInfo = await statusResponse.json();
            console.log('Video info:', videoInfo);
            
            if (videoInfo.status !== 'ready') {
                throw new Error(`Video is not ready (status: ${videoInfo.status})`);
            }
            
            const defaultQuality = videoInfo.qualities.find(q => q.resolution === '720p') || videoInfo.qualities[0];
            if (!defaultQuality) {
                throw new Error('No video qualities available');
            }
            
            const videoPath = `/api/videos/${videoId}/stream?quality=${defaultQuality.resolution}`;
            console.log('Setting video source:', videoPath);
            
            // 添加时间戳防止缓存
            const timestamp = new Date().getTime();
            const videoUrl = `${videoPath}&t=${timestamp}`;
            
            // 先检查视频是否可访问
            const checkResponse = await fetch(videoUrl, { method: 'HEAD' });
            if (!checkResponse.ok) {
                throw new Error(`Video file not accessible: ${checkResponse.status}`);
            }
            
            // 设置视频源
            this.player.src = videoUrl;
            this.player.load();
            
            // 设置质量选择器
            this.setupQualitySelector(videoInfo.qualities);
            
        } catch (error) {
            console.error('Failed to load video:', error);
            alert('Failed to load video: ' + error.message);
        }
    }
    
    retryLoad() {
        const currentSrc = this.player.src;
        if (currentSrc) {
            // 添加时间戳防止缓存
            const timestamp = new Date().getTime();
            const newSrc = currentSrc.includes('?') 
                ? `${currentSrc.split('&t=')[0]}&t=${timestamp}`
                : `${currentSrc}?t=${timestamp}`;
            
            console.log('Retrying with new source:', newSrc);
            this.player.src = newSrc;
            this.player.load();
        }
    }
    
    setupQualitySelector(qualities) {
        // 移除现有的质量选择器（如果存在）
        const existingSelector = this.player.parentElement.querySelector('.quality-selector');
        if (existingSelector) {
            existingSelector.remove();
        }

        const container = document.createElement('div');
        container.className = 'quality-selector';
        
        qualities.forEach(quality => {
            const button = document.createElement('button');
            button.textContent = quality.resolution;
            button.onclick = () => this.changeQuality(quality);
            container.appendChild(button);
        });
        
        this.player.parentElement.appendChild(container);
    }
    
    changeQuality(quality) {
        const currentTime = this.player.currentTime;
        const videoId = this.player.src.split('/')[4];
        const newPath = `/api/videos/${videoId}/stream?quality=${quality.resolution}`;
        this.player.src = newPath;
        this.player.currentTime = currentTime;
        this.player.play().catch(error => {
            console.error('Play failed after quality change:', error);
        });
    }
}

// 不要在这里初始化播放器
window.VideoPlayer = VideoPlayer; 