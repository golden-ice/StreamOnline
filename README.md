# Video Streaming Platform

A modern video streaming platform built with Go and JavaScript, supporting video upload, transcoding, and adaptive streaming.

## Features

- Video upload with chunked transfer support
- Automatic video transcoding to multiple resolutions (1080p, 720p, 480p)
- Adaptive streaming with HLS
- Modern web interface with Tailwind CSS
- Real-time upload progress tracking
- Video library management
- Automatic cleanup of temporary files

## Prerequisites

- Go 1.16 or higher
- FFmpeg (for video transcoding)
- Modern web browser with JavaScript enabled

## Installation

1. Clone the repository:
```bash
cd StreamOnline
```

2. Install FFmpeg:
- Windows: Download from [FFmpeg website](https://ffmpeg.org/download.html)
- Linux: `sudo apt-get install ffmpeg`
- macOS: `brew install ffmpeg`

3. Install Go dependencies:
```bash
go mod download
```

## Project Structure
.
├── frontend/ # Frontend files
│ ├── index.html # Main page
│ └── js/ # JavaScript files
├── handlers/ # HTTP handlers
├── models/ # Database models
├── services/ # Business logic
├── videos/ # Video storage
│ └── temp/ # Temporary upload directory
└── main.go # Application entry point


## Configuration

The application uses the following default settings:
- Server port: 8080
- Video storage: ./videos
- Temporary files: ./videos/temp
- Database: ./videos.db

## Usage

1. Start the server:
```bash
go run main.go
```

2. Open your browser and navigate to:
http://localhost:8080

3. Upload a video:
   - Click "Choose File" to select a video
   - Click "Upload" to start the upload process
   - Wait for the upload and transcoding to complete (Note: Restart the server and press F5 to refresh the html page)
   - The video will appear in your library

## API Endpoints

- `POST /api/upload/init` - Initialize upload
- `POST /api/upload/chunk` - Upload video chunk
- `POST /api/upload/complete` - Complete upload
- `GET /api/videos` - Get video list
- `GET /api/videos/:id` - Get video info
- `GET /api/videos/:id/stream` - Stream video

## Maintenance

The system automatically:
- Cleans up temporary files older than 24 hours
- Verifies video file integrity
- Removes invalid video entries from the database

## Notes

- After uploading a video, the server needs to be restarted to make the video available for playback
- Supported video formats: MP4, MOV, AVI
- Maximum file size: 2GB
- Recommended video codec: H.264

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

MIT License

Copyright (c) 2025 Bin Yan

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.