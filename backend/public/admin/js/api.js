/* API Client for Communicating with the Backend */

const API_CLIENT_BASE = `${API_BASE_URL}/admin`;

/**
 * Make an HTTP request to the API
 * @param {string} endpoint - The API endpoint (without base URL)
 * @param {Object} options - Fetch options
 * @returns {Promise<Response>}
 */
async function apiRequest(endpoint, options = {}) {
    const url = `${API_CLIENT_BASE}${endpoint}`;

    // Default headers
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };
    const token = localStorage.getItem('jwt') || sessionStorage.getItem('jwt');
    if (token && !headers.Authorization) {
        headers.Authorization = `Bearer ${token}`;
    }

    // Include credentials for cookies (session)
    const fetchOptions = {
        ...options,
        headers,
        credentials: 'include'
    };

    try {
        const response = await fetch(url, fetchOptions);
        return response;
    } catch (error) {
        console.error(`API request to ${endpoint} failed:`, error);
        throw error;
    }
}

/**
 * GET request
 * @param {string} endpoint
 * @param {Object} params - Query parameters
 * @returns {Promise<Response>}
 */
async function get(endpoint, params = {}) {
    // Build query string
    const queryString = new URLSearchParams(params).toString();
    const url = `${endpoint}${queryString ? '?' + queryString : ''}`;
    return await apiRequest(url, { method: 'GET' });
}

/**
 * POST request
 * @param {string} endpoint
 * @param {Object} data - Request body
 * @returns {Promise<Response>}
 */
async function post(endpoint, data = {}) {
    return await apiRequest(endpoint, {
        method: 'POST',
        body: JSON.stringify(data)
    });
}

/**
 * PUT request
 * @param {string} endpoint
 * @param {Object} data - Request body
 * @returns {Promise<Response>}
 */
async function put(endpoint, data = {}) {
    return await apiRequest(endpoint, {
        method: 'PUT',
        body: JSON.stringify(data)
    });
}

/**
 * DELETE request
 * @param {string} endpoint
 * @returns {Promise<Response>}
 */
async function del(endpoint) {
    return await apiRequest(endpoint, {
        method: 'DELETE'
    });
}

/**
 * Upload file (for future use)
 * @param {string} endpoint
 * @param {FormData} formData
 * @returns {Promise<Response>}
 */
async function uploadFile(endpoint, formData) {
    const url = `${API_CLIENT_BASE}${endpoint}`;
    const token = localStorage.getItem('jwt') || sessionStorage.getItem('jwt');
    const headers = token ? { Authorization: `Bearer ${token}` } : {};
	try {
		const response = await fetch(url, {
			method: 'POST',
			headers,
			body: formData,
			credentials: 'include'
		});
        return response;
    } catch (error) {
        console.error(`File upload to ${endpoint} failed:`, error);
        throw error;
    }
}

// Export functions
window.api = {
    get,
    post,
    put,
	delete: del,
	uploadFile,
	API_BASE: API_CLIENT_BASE
};
