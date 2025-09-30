# URL Shortener Frontend

Vue 3 frontend application for testing and managing the URL Shortener service.

## Features

- ✂️ **URL Shortening**: Shorten URLs with a simple interface
- 📊 **Analytics Dashboard**: View detailed analytics for shortened URLs
- 🏥 **Health Check**: Monitor system health and performance
- ⚡ **Load Testing**: Perform load tests on the API
- 📦 **Bulk Operations**: Shorten multiple URLs and get analytics in bulk

## Tech Stack

- **Vue 3** with Composition API
- **TypeScript** for type safety
- **Vite** for fast development and building
- **Axios** for API communication
- **CSS3** with modern styling and animations

## Development

### Prerequisites

- Node.js 18+ 
- npm or yarn

### Installation

```bash
# Install dependencies
npm install

# Copy environment file
cp env.example .env

# Start development server
npm run dev
```

The application will be available at `http://localhost:3000`

### Available Scripts

```bash
# Development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint code
npm run lint
```

## Production Deployment

### Docker

```bash
# Build Docker image
docker build -t url-shortener-frontend .

# Run container
docker run -p 80:80 url-shortener-frontend
```

### Environment Variables

- `VITE_API_URL`: Backend API URL (default: http://localhost:8080)
- `VITE_ENV`: Environment (development/production)

## API Integration

The frontend communicates with the URL Shortener backend API:

- `POST /api/v1/shorten` - Shorten URL
- `GET /api/v1/analytics/:shortCode` - Get analytics
- `GET /api/v1/health` - Health check
- `GET /:shortCode` - Redirect to original URL

## Components

### UrlShortener
- Shorten URLs
- Copy results to clipboard
- Test redirect functionality

### Analytics
- View detailed analytics
- Click statistics and charts
- Top referrers and user agents
- Daily statistics

### HealthCheck
- System health monitoring
- Database and Redis status
- Performance metrics
- Auto-refresh capability

### LoadTesting
- Configurable load tests
- Real-time progress tracking
- Performance statistics
- Error analysis

### BulkOperations
- Bulk URL shortening
- Bulk analytics retrieval
- CSV export functionality
- Progress tracking

## Styling

The application uses modern CSS with:
- Responsive design
- Gradient backgrounds
- Smooth animations
- Card-based layout
- Mobile-first approach

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License
