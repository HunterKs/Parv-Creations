/* Authentication Helper Functions */

const API_BASE = '/api/v1/admin';

/**
 * Check if the user is authenticated by looking for a valid session token
 * @returns {Promise<boolean>}
 */
async function isAuthenticated() {
    try {
        const response = await fetch(`${API_BASE}/auth/status`, {
            method: 'GET',
            credentials: 'include' // Important for sending cookies
        });
        return response.ok;
    } catch (error) {
        console.error('Auth check failed:', error);
        return false;
    }
}

/**
 * Get the current user's information from the session
 * @returns {Promise<Object|null>}
 */
async function getCurrentUser() {
    try {
        const response = await fetch(`${API_BASE}/auth/me`, {
            method: 'GET',
            credentials: 'include'
        });
        if (!response.ok) return null;
        return await response.json();
    } catch (error) {
        console.error('Failed to get current user:', error);
        return null;
    }
}

/**
 * Log out the current user
 * @returns {Promise<boolean>}
 */
async function logout() {
    try {
        const response = await fetch(`${API_BASE}/auth/logout`, {
            method: 'POST',
            credentials: 'include'
        });
        // Clear any local storage if needed
        localStorage.removeItem('token');
        sessionStorage.removeItem('token');
        return response.ok;
    } catch (error) {
        console.error('Logout failed:', error);
        return false;
    }
}

/**
 * Set up automatic redirect to login if not authenticated
 * @param {string} loginUrl - URL to redirect to when not authenticated
 */
function setupAuthGuard(loginUrl = '/admin/login.html') {
    // Check authentication on page load for protected pages
    if (window.location.pathname !== loginUrl && !window.location.pathname.includes('login')) {
        isAuthenticated().then(authenticated => {
            if (!authenticated) {
                // Redirect to login page
                window.location.href = loginUrl;
            }
        });
    }

    // Also check periodically or on visibility change if needed
}

/**
 * Attach logout event listener to elements with data-logout attribute
 */
function setupLogoutListeners() {
    document.addEventListener('click', async (e) => {
        if (e.target.matches('[data-logout]')) {
            e.preventDefault();
            const success = await logout();
            if (success) {
                // Redirect to login page
                window.location.href = '/admin/login.html';
            } else {
                showNotification('Logout failed', 'error');
            }
        }
    });
}

/**
 * Initialize auth-related functionality
 */
function initAuth() {
    setupLogoutListeners();
    // We don't automatically redirect here because some pages (like login) should be accessible
    // Individual pages can call setupAuthGuard() if they need protection
}

// Export functions for use in other scripts
window.auth = {
    isAuthenticated,
    getCurrentUser,
    logout,
    setupAuthGuard,
    setupLogoutListeners,
    initAuth
};

// Initialize on DOM content loaded
document.addEventListener('DOMContentLoaded', initAuth);