# VIDI (Data Visualization System)

**MIT License © Matthew Salvatore Giancola**

## Overview

VIDI is a data management and visualization system that enables visualization of data put into HTML forms as cards. It also manages cookies, secrets and other web-based data with built-in security features. It can automatically paginate large loading data and integrate URL-based sources with CORS handling.

## Installation

### Prerequisites
- Node.js (for build tools, optional for runtime)
- Modern web browser with JavaScript support
- No external server dependencies required

### Installation Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/emperor42/vidi.git
   cd vidi
   ```

2. Include VIDI in your HTML:
   ```html
   <!-- Option 1: Direct script tag -->
   <script src="vidi.js"></script>

   <!-- Option 2: Local file -->
   <script>
   // VIDI will automatically initialize
   </script>
   ```

3. For Node.js development (optional):
   ```bash
   # Install build tools
   npm install
   
   # Build for distribution
   npm run build
   ```

## Usage (Standalone)

### Basic Usage

```javascript
// Basic VIDI usage
<script src="vidi.js"></script>

// VIDI will automatically initialize with default configuration
const vidi = new VIDI({
  dataSource: 'https://api.example.com/data',
  pagination: { pageSize: 10, enabled: true },
  enableCORS: true,
  cookieSettings: { secure: false, httpOnly: true }
});

// Initialize and render
vidi.init();
```

### Advanced Usage

```javascript
// Create a VIDI instance with advanced configuration
const vidi = new VIDI({
  dataSource: 'https://api.example.com/data',
  encryptionKey: 'your-encryption-key',
  pagination: { pageSize: 20, enabled: true },
  enableCORS: true,
  cookieSettings: { secure: true, httpOnly: true },
  onDataLoad: function(data) {
    console.log(`Loaded ${data.length} data items`);
  },
  onCardSelect: function(item) {
    console.log(`Selected card: ${item.title}`);
  }
});

// Load data from source
vidi.loadDataFromSource().then(() => {
  // Render data as cards
  vidi.renderData();
});\n// Query data
const filteredData = vidi.queryData({
  category: 'important',
  search: 'urgent'
});

// Get paginated data
const pageData = vidi.getPaginatedData(1, { category: 'recent' });
```

### API Endpoints

| Method | Description |
|--------|-------------|
| `new VIDI(options)` | Create new VIDI instance |
| `vidi.init()` | Initialize VIDI |
| `vidi.loadData()` | Load data from source or storage |
| `vidi.addData(data)` | Add new data item |
| `vidi.updateData(id, updates)` | Update existing data item |
| `vidi.deleteData(id)` | Delete data item |
| `vidi.queryData(filters)` | Query data with filters |
| `vidi.getPaginatedData(page, filters)` | Get paginated data |
| `vidi.getTotalPages(filters)` | Get total pages for pagination |
| `vidi.renderData()` | Render data as cards |
| `vidi.renderCards(data)` | Render specific data as cards |
| `vidi.createCard(item)` | Create card element for item |
| `vidi.setupPaginationControls()` | Setup pagination controls |

### Example HTML Page

```html
<!DOCTYPE html>
<html>
<head>
    <title>VIDI Data Visualization Demo</title>
    <!-- Include VIDI -->
    <script src="vidi.js"></script>
</head>
<body>
    <!-- VIDI will automatically create card visualization -->
    <div id="vidi-container"></div>
    <!-- VIDI will render data as cards here -->
    
    <!-- Pagination controls -->
    <div id="pagination"></div>
    
    <script>
    // VIDI automatically initializes after loading
    document.addEventListener('VIDI-ready', function() {
        // VIDI has loaded and rendered data
        console.log('VIDI is ready');
        
        // Example: Handle card selection
        vidi.on('vidi:card-select', (event) => {
            const { item } = event.detail;
            console.log('Card selected:', item);
            // Handle card selection
        });
    });
    </script>
</body>
</html>
```

## Integration with ATP

### Data Integration

VIDI integrates with ATP to provide centralized data management and visualization:

```javascript
// VIDI with ATP integration
const vidi = new VIDI({
    dataSource: '/atp/api/data',
    pagination: { pageSize: 10, enabled: true },
    enableCORS: true,
    cookieSettings: { secure: false, httpOnly: true },
    onDataLoad: function(data) {
        console.log(`Loaded ${data.length} data items from ATP`);
        vidi.renderData();
    },
    onError: function(error) {
        console.error('Data loading error:', error);
    }
});

// Load data from ATP
vidi.loadDataFromSource().then(() => {
    console.log('Data loaded from ATP successfully');
});
```

### Configuration Integration

```javascript
// VIDI configuration for ATP integration
const vidiConfig = {
    dataSource: '/atp/api/data',
    pagination: { pageSize: 10, enabled: true },
    enableCORS: true,
    cookieSettings: { secure: false, httpOnly: true },
    encryptionKey: 'your-encryption-key',
    apiEndpoint: '/atp/api',
    syncStrategy: 'merge', // merge, replace, append
    onDataLoad: function(data) {
        console.log(`Loaded ${data.length} data items`);
        vidi.renderData();
    },
    onCardSelect: function(item) {
        console.log(`Card selected: ${item.title}`);
        // Handle card selection
    },
    onError: function(error) {
        console.error('VIDI error:', error);
    }
};
```

### Data Management Pipeline

1. **Data Collection**: VIDI collects data from ATP or configured sources
2. **Data Processing**: Raw data is processed and normalized
3. **Data Visualization**: Processed data is visualized as cards
4. **Data Storage**: Data is stored in local storage or cookies
5. **Data Distribution**: Data is distributed through VIDI APIs
6. **User Interaction**: Users interact with visualized data

## Development Setup

### Local Development

```bash
# Test in browser
# Open browser and load:
# http://localhost:8080/vidi.html
# (VIDI will automatically load and display data)

# Or with Node.js
node -e "require('vidi').test()"
```

### Testing

```javascript
// Basic VIDI usage test
const VIDI = window.VIDI;

const testVidi = () => {
    // Test VIDI initialization
    const vidi = new VIDI({
        pagination: { pageSize: 10, enabled: true },
        enableCORS: true
    });
    
    expect(vidi).toBeDefined();
    expect(vidi.pagination).toBeDefined();
    
    // Test data loading
    vidi.init();
    const data = vidi.getPaginatedData();
    expect(data).toBeDefined();
};

// Data visualization test
const testDataVisualization = () => {
    const vidi = new VIDI({
        pagination: { pageSize: 10, enabled: true },
        enableCORS: true
    });
    
    vidi.init();
    
    // Mock data
    const mockData = [
        { id: 1, title: 'Card 1', description: 'First card', category: 'important' },
        { id: 2, title: 'Card 2', description: 'Second card', category: 'normal' },
        { id: 3, title: 'Card 3', description: 'Third card', category: 'urgent' }
    ];
    
    // Add mock data
    mockData.forEach(item => vidi.addData(item));
    
    // Render data
    vidi.renderData();
    
    // Check if data was rendered
    const cards = document.querySelectorAll('.vidi-card');
    expect(cards.length).toBe(mockData.length);
};
```

### Building

```bash
# Build for distribution
npm run build

# Output: dist/vidi.js (optimized and bundled)

# Test in browser
# Open browser and load: dist/vidi.js
```

## API Specifications

### High Maturity API (Event-driven)

#### Data Management
- `new VIDI(options)` - Create new VIDI instance
- `vidi.init()` - Initialize VIDI
- `vidi.loadData()` - Load data from source or storage
- `vidi.addData(data)` - Add new data item
- `vidi.updateData(id, updates)` - Update existing data item
- `vidi.deleteData(id)` - Delete data item
- `vidi.queryData(filters)` - Query data with filters
- `vidi.getPaginatedData(page, filters)` - Get paginated data
- `vidi.getTotalPages(filters)` - Get total pages for pagination

#### Data Visualization
- `vidi.renderData()` - Render data as cards
- `vidi.renderCards(data)` - Render specific data as cards
- `vidi.createCard(item)` - Create card element for item
- `vidi.setupPaginationControls()` - Setup pagination controls

#### Storage Management
- `vidi.saveToStorage(data)` - Save data to local storage
- `vidi.loadFromStorage()` - Load data from local storage
- `vidi.clearData()` - Clear all data

#### Security and Authentication
- `vidi.hashPassword(password)` - Hash password using SHA-256
- `vidi.validatePassword(password, storedHash)` - Validate password against stored hash
- `vidi.manageCookies(action, name, value, options)` - Manage cookies
- `vidi.encryptData(data)` - Encrypt data with AES
- `vidi.decryptData(encryptedData)` - Decrypt data

#### Event Handling
- `vidi.on(event, callback)` - Listen for VIDI events

### VIDI APIs

```javascript
// Create VIDI instance
const vidi = new VIDI({
  dataSource: 'https://api.example.com/data',
  pagination: { pageSize: 10, enabled: true },
  enableCORS: true,
  cookieSettings: { secure: false, httpOnly: true }
});

// Load data
await vidi.loadData();

// Render data
vidi.renderData();

// Add new data
vidi.addData({
  id: 4,
  title: 'Card 4',
  description: 'Fourth card',
  category: 'normal'
});

// Query data
const filteredData = vidi.queryData({
  category: 'important',
  search: 'urgent'
});

// Get paginated data
const pageData = vidi.getPaginatedData(1, { category: 'recent' });

// Event handling
vidi.on('vidi:add', (event) => {
  console.log('Data added:', event.detail.data);
});

vidi.on('vidi:update', (event) => {
  console.log('Data updated:', event.detail);
});

vidi.on('vidi:delete', (event) => {
  console.log('Data deleted:', event.detail.id);
});
```

### VIDI-specific Events

```javascript
// Data add event
vidi.on('vidi:add', (event) => {
  const { detail } = event;
  vidi.addData(detail.data);
});

// Data update event
vidi.on('vidi:update', (event) => {
  const { detail } = event;
  vidi.updateData(detail.id, detail.updates);
});

// Data delete event
vidi.on('vidi:delete', (event) => {
  const { detail } = event;
  vidi.deleteData(detail.id);
});

// Card select event
vidi.on('vidi:card-select', (event) => {
  const { item } = event.detail;
  console.log('Card selected:', item);
});
```

## Security API

### Data Security
- `vidi.encryptData(data)` - Encrypt data with AES
- `vidi.decryptData(encryptedData)` - Decrypt data
- `vidi.hashPassword(password)` - Hash password using SHA-256
- `vidi.validatePassword(password, storedHash)` - Validate password

### Cookie Management
- `vidi.manageCookies(action, name, value, options)` - Manage cookies
- `vidi.buildCookieString(name, value, options)` - Build cookie string
- `vidi.parseCookie(name)` - Parse cookie value

### CORS Configuration
- `vidi.fetchData(url)` - Fetch data with CORS support
- `vidi.setCORSOrigin(origin)` - Set CORS origin
- `vidi.getCORSPolicy()` - Get CORS policy

## Integration API

### VENI Integration
- `vidi.discoverVENIComponents()` - Discover VENI components
- `vidi.integrateWithVENI(component)` - Integrate with VENI component
- `vidi.generateFromVENITemplate(template)` - Generate from VENI template

### VICI Integration
- `vidi.syncWithVICI(changes)` - Synchronize with VICI
- `vidi.getVIDIChanges()` - Get VICI changes
- `vidi.applyVIDIChanges(changes)` - Apply VICI changes

### VINI Integration
- `vidi.getVINIWorkflows()` - Get VINI workflows
- `vidi.integrateWithVINI(workflow)` - Integrate with VINI workflow
- `vidi.executeVINIWorkflow(workflow)` - Execute VINI workflow

## Monitoring API

### Data Monitoring
- `vidi.getTotalItems()` - Get total data items
- `vidi.getFilteredDataCount(filters)` - Get filtered data count
- `vidi.getPaginationInfo(page, filters)` - Get pagination information

### Performance Monitoring
- `vidi.getPerformanceMetrics()` - Get performance metrics
- `vidi.getDataLoadTime()` - Get data load time
- `vidi.getRenderTime()` - Get render time

### Event Monitoring
- `vidi.onDataLoad(callback)` - Data load callback
- `vidi.onDataError(callback)` - Data error callback
- `vidi.onCardRender(callback)` - Card render callback

## Error Handling

### VIDI Error Types
- `DataError` - Data loading and processing errors
- `SecurityError` - Security-related errors
- `ValidationError` - Input validation errors
- `IntegrationError` - Integration-related errors

### Error Response Format
```javascript
// VIDI errors
class VIDISecurityError extends Error {
  constructor(message, code, details) {
    super(message);
    this.code = code;
    this.details = details;
    this.timestamp = new Date().toISOString();
  }
}
```

## Testing

### Unit Tests

```javascript
// Test data loading
 test('Data Loading', () => {
   const vidi = new VIDI({
     dataSource: 'test-data.json'
   });
   vidi.init();
   expect(vidi.getTotalItems()).toBeGreaterThan(0);
 });

// Test data querying
 test('Data Querying', () => {
   const vidi = new VIDI();
   const filteredData = vidi.queryData({ category: 'test' });
   expect(filteredData).toBeDefined();
 });
```

### Integration Tests

```javascript
// Test VIDI integration
 test('VIDI-VENI Integration', () => {
   const vidi = new VIDI({
     dataSource: 'https://venisite.com/data'
   });
   vidi.init();
   expect(vidi.getTotalItems()).toBeGreaterThan(0);
 });

// Test VIDI-VICI integration
 test('VIDI-VICI Integration', () => {
   const vidi = new VIDI();
   const changes = vidi.getVIDIChanges();
   expect(changes).toBeDefined();
 });
```

## Performance Considerations

- **Memory Usage**: Monitor for large datasets
- **CPU Usage**: Optimize data processing algorithms
- **Network I/O**: Cache frequently accessed data
- **Disk I/O**: Use efficient storage for large datasets
- **Concurrent Processing**: Support for concurrent data operations

## Future Enhancements

- **Advanced Analytics**: Add advanced data analytics
- **Machine Learning**: ML-powered data analysis
- **Real-time Processing**: Add real-time data updates
- **Advanced Visualization**: Enhanced visualization capabilities
- **Cloud Integration**: Integrate with cloud data services

## Troubleshooting

### Common Issues

1. **Data not loading**
   ```javascript
   // Check VIDI configuration
   const vidi = new VIDI({
     dataSource: 'https://api.example.com/data',
     pagination: { pageSize: 10, enabled: true }
   });
   
   vidi.init();
   
   // Check if data was loaded
   const totalItems = vidi.getTotalItems();
   console.log(`Total items: ${totalItems}`);
   ```

2. **CORS errors**
   ```javascript
   // Check CORS configuration
   const vidi = new VIDI({
     dataSource: 'https://api.example.com/data',
     enableCORS: true
   });
   
   // Test CORS handling
   vidi.fetchData('https://api.example.com/data').then(data => {
     console.log('CORS data loaded:', data);
   });
   ```

3. **Performance issues**
   ```javascript
   // Enable performance monitoring
   vidi.setPerformanceMonitoring(true);
   
   // Check performance metrics
   const metrics = vidi.getPerformanceMetrics();
   console.log('Performance metrics:', metrics);
   ```

### Debugging Commands

```javascript
// Enable debug logging
vidi.setDebugMode(true);

// Check data logs
const logs = vidi.getDataLogs();
console.log(logs);

// Monitor system resources
// Use browser dev tools to monitor performance
```

## Conclusion

VIDI provides a comprehensive data management and visualization solution that enables users to visualize data in an intuitive card-based format while providing advanced data security and management features. It offers rich integration capabilities with other Emperor42 projects and can be easily embedded in web applications.

Key benefits:

- **Data Visualization**: Card-based data visualization
- **Data Management**: Comprehensive data storage and management
- **Security Features**: Built-in security and authentication
- **CORS Support**: Cross-origin resource sharing
- **Pagination**: Automatic pagination for large datasets
- **Integration**: Rich integration with VENI, VICI, and VINI
- **Monitoring**: Comprehensive data monitoring and analytics
- **Flexibility**: Flexible data management and visualization

This data management system is production-ready and can be easily integrated into web applications with comprehensive data visualization and security features.

---

*Document Version: 1.0*
*Created: 2026-08-25*
*Last Updated: 2026-08-25*
*Status: Production Ready*

**License:** MIT License © Matthew Salvatore Giancola.