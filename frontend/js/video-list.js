class VideoList {
    constructor() {
        this.videoList = document.getElementById('videoList');
        this.player = new VideoPlayer();
        this.loadVideos();
    }

    async loadVideos() {
        try {
            console.log('Loading videos...');
            const response = await fetch('/api/videos');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const videos = await response.json();
            console.log('Loaded videos:', videos);
            
            // 检查视频状态
            videos.forEach(video => {
                console.log(`Video ${video.id} status: ${video.status}`);
            });
            
            this.renderVideos(videos);
        } catch (error) {
            console.error('Failed to load videos:', error);
            this.videoList.innerHTML = '<div class="error">Failed to load videos</div>';
        }
    }

    renderVideos(videos) {
        console.log('Rendering videos:', videos);
        this.videoList.innerHTML = '';
        
        if (!videos || videos.length === 0) {
            this.videoList.innerHTML = `
                <div class="col-span-full text-center py-8 text-gray-500">
                    No videos available
                </div>`;
            return;
        }
        
        videos.forEach(video => {
            console.log(`Rendering video: ${video.id}, status: ${video.status}`);
            const videoCard = document.createElement('div');
            videoCard.className = `video-card bg-white rounded-lg shadow-md overflow-hidden
                                 ${video.status === 'ready' ? 'cursor-pointer hover:shadow-lg transition-shadow duration-200' : 'opacity-50'}`;
            
            if (video.status === 'ready') {
                videoCard.onclick = () => this.playVideo(video);
            } else {
                videoCard.title = `Video status: ${video.status}`;
            }

            // Format status for display
            const statusDisplay = {
                'ready': 'Ready',
                'processing': 'Processing',
                'error': 'Error',
                'pending': 'Pending'
            }[video.status] || video.status;

            videoCard.innerHTML = `
                <div class="relative aspect-w-16 aspect-h-9">
                    <img src="/static/images/video-placeholder.png" alt="Video thumbnail"
                         class="w-full h-full object-cover">
                    ${video.status !== 'ready' ? `
                        <div class="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center">
                            <span class="text-white text-sm">${statusDisplay}</span>
                        </div>
                    ` : ''}
                </div>
                <div class="p-4">
                    <h3 class="text-lg font-semibold text-gray-800 truncate">${video.title}</h3>
                    <p class="text-sm text-gray-500 mt-1">
                        ${new Date(video.createdAt).toLocaleDateString()}
                    </p>
                    <div class="mt-2 flex items-center space-x-2">
                        <span class="px-2 py-1 text-xs rounded-full
                                   ${video.status === 'ready' ? 'bg-green-100 text-green-800' : 
                                     video.status === 'processing' ? 'bg-yellow-100 text-yellow-800' : 
                                     'bg-red-100 text-red-800'}">
                            ${statusDisplay}
                        </span>
                    </div>
                </div>
            `;

            this.videoList.appendChild(videoCard);
        });
    }

    playVideo(video) {
        console.log('Playing video:', video);
        if (!video || !video.id) {
            console.error('Invalid video object:', video);
            return;
        }
        if (video.status !== 'ready') {
            console.error('Video is not ready:', video.status);
            return;
        }
        this.player.loadVideo(video.id);
        document.querySelector('.player-section').scrollIntoView({ behavior: 'smooth' });
    }
}

// 初始化视频列表
document.addEventListener('DOMContentLoaded', () => {
    window.videoList = new VideoList();
}); 