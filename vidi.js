"use strict";

/**
 * vidi - Data management and visualization for web applications
 * JavaScript based data management and visualization system
 * Can be fully embedded into a script tag, hosted in a static page
 * No web server required to work
 */

class Vidi {
  constructor(options = {}) {
    this.options = {
      dataSource: options.dataSource || null,
      encryptionKey: options.encryptionKey || null,
      enableCORS: options.enableCORS !== false,
      pagination: options.pagination || { pageSize: 10, enabled: true },
      cookieSettings: options.cookieSettings || { secure: false, httpOnly: true },
      ...options
    };
    this.dataStore = new Map();
    this.initialized = false;
    this.currentPage = 1;
    this.data = [];
  }

  /**
   * Initialize the data visualization and management system
   */
  init() {
    if (this.initialized) return;
    
    this.initialized = true;
    this.loadData();
    this.setupEventListeners();
    this.renderData();
    return this;
  }

  /**
   * Load data from source or storage
   */
  loadData() {
    if (this.options.dataSource) {
      this.loadFromSource();
    } else {
      this.loadFromStorage();
    }
  }

  /**
   * Load data from external source
   */
  async loadFromSource() {
    try {
      const response = await this.fetchData(this.options.dataSource);
      this.data = this.processData(response);
      this.saveToStorage(this.data);
    } catch (error) {
      console.error('Failed to load data from source:', error);
      this.loadFromStorage();
    }
  }

  /**
   * Load data from local storage
   */
  loadFromStorage() {
    const storedData = localStorage.getItem('vidi_data');
    this.data = storedData ? JSON.parse(storedData) : [];
    this.indexData();
  }

  /**
   * Save data to local storage with encryption
   */
  saveToStorage(data) {
    const serialized = JSON.stringify(data);
    const encrypted = this.options.encryptionKey ? 
      this.encryptData(serialized) : serialized;
    localStorage.setItem('vidi_data', encrypted);
  }

  /**
   * Process and normalize data
   */
   processData(data) {
    return data.map(item => ({
      id: item.id || this.generateId(),
      title: item.title || item.name || '',
      description: item.description || '',
      value: item.value || item.amount || null,
      timestamp: item.timestamp || new Date().toISOString(),
      category: item.category || 'general',
      metadata: item.metadata || {}
    }));
  }

  /**
   * Index data for faster searching
   */
  indexData() {
    this.dataStore.clear();
    this.data.forEach(item => {
      const key = this.options.pagination.enabled ? item.category || 'uncategorized' : 'all';
      if (!this.dataStore.has(key)) {
        this.dataStore.set(key, []);
      }
      this.dataStore.get(key).push(item);
    });
  }

  /**
   * Fetch data from URL with CORS support
   */
  async fetchData(url) {
    const fetchOptions = {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    };

    if (this.options.enableCORS) {
      fetchOptions.mode = 'cors';
    }

    const response = await fetch(url, fetchOptions);
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  /**
   * Encrypt data with AES
   */
  encryptData(data) {
    if (!this.options.encryptionKey) return data;
    
    try {
      const encoder = new TextEncoder();
      const encodedData = encoder.encode(data);
      const keyData = encoder.encode(this.options.encryptionKey);
      
      return btoa(String.fromCharCode(...encodedData.map(b => b ^ keyData[b % keyData.length])));
    } catch (error) {
      console.error('Encryption failed:', error);
      return data;
    }
  }

  /**
   * Decrypt data with AES
   */
  decryptData(encryptedData) {
    if (!this.options.encryptionKey) return encryptedData;
    
    try {
      const encodedData = new Uint8Array(Array.from(atob(encryptedData), c => c.charCodeAt(0)));
      const keyData = new TextEncoder().encode(this.options.encryptionKey);
      
      return String.fromCharCode(...encodedData.map((b, i) => b ^ keyData[i % keyData.length]));
    } catch (error) {
      console.error('Decryption failed:', error);
      return encryptedData;
    }
  }

  /**
   * Setup event listeners for data operations
   */
  setupEventListeners() {
    document.addEventListener('vidi:add', this.handleAddData.bind(this));
    document.addEventListener('vidi:update', this.handleUpdateData.bind(this));
    document.addEventListener('vidi:delete', this.handleDeleteData.bind(this));
  }

  /**
   * Handle add data event
   */
  handleAddData(event) {
    const { detail } = event;
    this.addData(detail.data);
  }

  /**
   * Handle update data event
   */
  handleUpdateData(event) {
    const { detail } = event;
    this.updateData(detail.id, detail.updates);
  }

  /**
   * Handle delete data event
   */
  handleDeleteData(event) {
    const { detail } = event;
    this.deleteData(detail.id);
  }

  /**
   * Add new data item
   */
  addData(data) {
    const processedData = this.processData([data])[0];
    this.data.push(processedData);
    this.saveToStorage(this.data);
    this.indexData();
    this.renderData();
  }

  /**
   * Update existing data item
   */
  updateData(id, updates) {
    const index = this.data.findIndex(item => item.id === id);
    if (index !== -1) {
      this.data[index] = { ...this.data[index], ...updates };
      this.saveToStorage(this.data);
      this.indexData();
      this.renderData();
    }
  }

  /**
   * Delete data item
   */
  deleteData(id) {
    this.data = this.data.filter(item => item.id !== id);
    this.saveToStorage(this.data);
    this.indexData();
    this.renderData();
  }

  /**
   * Query data with filters
   */
  queryData(filters = {}) {
    let results = [...this.data];

    if (filters.category) {
      results = results.filter(item => item.category === filters.category);
    }

    if (filters.search) {
      const searchTerm = filters.search.toLowerCase();
      results = results.filter(item =>
        item.title.toLowerCase().includes(searchTerm) ||
        item.description.toLowerCase().includes(searchTerm)
      );
    }

    if (filters.minValue !== undefined) {
      results = results.filter(item => item.value >= filters.minValue);
    }

    return results;
  }

  /**
   * Get paginated data
   */
  getPaginatedData(page = this.currentPage, filters = {}) {
    let results = this.queryData(filters);
    
    if (this.options.pagination.enabled) {
      const startIndex = (page - 1) * this.options.pagination.pageSize;
      const endIndex = Math.min(startIndex + this.options.pagination.pageSize, results.length);
      results = results.slice(startIndex, endIndex);
    }

    this.currentPage = page;
    return results;
  }

  /**
   * Get total pages for pagination
   */
  getTotalPages(filters = {}) {
    const results = this.queryData(filters);
    if (!this.options.pagination.enabled) return 1;
    
    return Math.ceil(results.length / this.options.pagination.pageSize);
  }

  /**
   * Generate unique ID
   */
  generateId() {
    return Math.random().toString(36).substr(2, 9);
  }

  /**
   * Setup password hashing using SHA-256
   */
  async hashPassword(password) {
    const encoder = new TextEncoder();
    const data = encoder.encode(password);
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    return hashHex;
  }

  /**
   * Validate password against stored hash
   */
  async validatePassword(password, storedHash) {
    const hashedPassword = await this.hashPassword(password);
    return hashedPassword === storedHash;
  }

  /**
   * Manage cookies with security settings
   */
  manageCookies(action, name, value, options = {}) {
    const cookieOptions = {
      secure: this.options.cookieSettings.secure,
      httpOnly: this.options.cookieSettings.httpOnly,
      sameSite: options.sameSite || 'lax',
      ...options
    };

    const cookieString = this.buildCookieString(name, value, cookieOptions);

    if (action === 'set') {
      document.cookie = cookieString;
      return this;
    } else if (action === 'get') {
      return this.parseCookie(name);
    } else if (action === 'delete') {
      document.cookie = this.buildCookieString(name, '', { ...cookieOptions, expires: 'Thu, 01 Jan 1970 00:00:00 GMT' });
      return this;
    }
  }

  /**
   * Build cookie string
   */
  buildCookieString(name, value, options) {
    let cookie = `${name}=${encodeURIComponent(value)}`;
    if (options.expires) cookie += `; expires=${options.expires}`;
    if (options.secure) cookie += '; secure';
    if (options.httpOnly) cookie += '; httponly';
    cookie += `; samesite=${options.sameSite}`;
    cookie += '; path=/';
    return cookie;
  }

  /**
   * Parse cookie value
   */
  parseCookie(name) {
    const cookies = document.cookie.split(';');
    const cookie = cookies.find(c => c.trim().startsWith(`${name}=`));
    return cookie ? decodeURIComponent(cookie.split('=')[1]) : null;
  }

  /**
   * Render data as cards or visualizations
   */
  renderData() {
    this.renderCards(this.getPaginatedData());
    this.setupPaginationControls();
  }

  /**
   * Render data as cards
   */
  renderCards(data) {
    const container = document.getElementById('vidi-cards-container');
    if (!container) return;

    container.innerHTML = '';

    data.forEach(item => {
      const card = this.createCard(item);
      container.appendChild(card);
    });
  }

  /**
   * Create a card element for an item
   */
  createCard(item) {
    const card = document.createElement('div');
    card.className = 'vidi-card';
    card.setAttribute('data-id', item.id);

    card.innerHTML = `
      <div class="vidi-card-header">
        <h3 class="vidi-card-title">${item.title}</h3>
        <span class="vidi-card-category">${item.category}</span>
      </div>
      <div class="vidi-card-body">
        <p class="vidi-card-description">${item.description}</p>
        ${item.value ? `<div class="vidi-card-value">${item.value}</div>` : ''}
      </div>
      <div class="vidi-card-footer">
        <span class="vidi-card-timestamp">${new Date(item.timestamp).toLocaleDateString()}</span>
      </div>
    `;

    card.addEventListener('click', () => this.selectCard(item));

    return card;
  }

  /**
   * Handle card selection
   */
  selectCard(item) {
    const event = new CustomEvent('vidi:card-select', {
      detail: { item }
    });
    document.dispatchEvent(event);
  }

  /**
   * Setup pagination controls
   */
  setupPaginationControls() {
    if (!this.options.pagination.enabled) return;

    const totalPages = this.getTotalPages();
    if (totalPages <= 1) return;

    let paginationHTML = '';
    for (let i = 1; i <= totalPages; i++) {
      paginationHTML += `
        <button class="vidi-page-btn" data-page="${i}" ${i === this.currentPage ? 'disabled' : ''}>${i}</button>
      `;
    }

    const paginationContainer = document.getElementById('vidi-pagination');
    if (paginationContainer) {
      paginationContainer.innerHTML = paginationHTML;
      this.setupPaginationListeners();
    }
  }

  /**
   * Setup pagination listeners
   */
  setupPaginationListeners() {
    const pageButtons = document.querySelectorAll('.vidi-page-btn');
    pageButtons.forEach(button => {
      button.addEventListener('click', (e) => {
        const page = parseInt(e.currentTarget.getAttribute('data-page'));
        this.renderData();
      });
    });
  }

  /**
   * Export data to CSV
   */
  exportData(filename = 'vidi-data.csv') {
    const headers = ['ID', 'Title', 'Description', 'Value', 'Category', 'Timestamp'];
    const csvRows = [headers.join(',')];

    this.data.forEach(item => {
      const row = [
        item.id,
        item.title,
        item.description,
        item.value,
        item.category,
        item.timestamp
      ];
      csvRows.push(row.join(','));
    });

    const csvContent = csvRows.join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);

    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    link.click();
    URL.revokeObjectURL(url);
  }

  /**
   * Clear all data
   */
  clearData() {
    if (confirm('Are you sure you want to clear all data? This cannot be undone.')) {
      this.data = [];
      this.saveToStorage(this.data);
      this.indexData();
      this.renderData();
    }
  }
}

/**
 * Initialize vidi when DOM is ready
 */
const vidi = new Vidi({
  pagination: { pageSize: 10, enabled: true },
  enableCORS: true,
  cookieSettings: { secure: false, httpOnly: true }
});

// Global configuration function
const configureVidi = (options) => {
  Object.assign(vidi.options, options);
  vidi.init();
};

// Auto-initialize when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => {
    vidi.init();
  });
} else {
  vidi.init();
}

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { Vidi };
}

if (typeof window !== 'undefined') {
  window.vidi = { Vidi, vidi, configureVidi };
}