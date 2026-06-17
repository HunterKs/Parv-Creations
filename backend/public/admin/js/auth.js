/* Authentication Helper Functions */

const AUTH_API_BASE = `${API_BASE_URL}/admin`;

/**
 * Check if the user is authenticated by looking for a valid session token
 * @returns {Promise<boolean>}
 */
async function isAuthenticated() {
    try {
        const response = await fetch(`${AUTH_API_BASE}/auth/status`, {
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
        const response = await fetch(`${AUTH_API_BASE}/auth/me`, {
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
        const response = await fetch(`${AUTH_API_BASE}/auth/logout`, {
            method: 'POST',
            credentials: 'include'
        });
        purgeLocalAuthState();
        return response.ok;
    } catch (error) {
        console.error('Logout failed:', error);
        purgeLocalAuthState();
        return false;
    }
}

function purgeLocalAuthState() {
    document.cookie = 'session_token=; Max-Age=0; path=/';
    document.cookie = 'jwt=; Max-Age=0; path=/';
    localStorage.clear();
    sessionStorage.clear();
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
            if (!success) showNotification('Logout failed', 'error');
            window.location.href = 'http://localhost:5500/admin/login.html';
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
    purgeLocalAuthState,
    setupAuthGuard,
    setupLogoutListeners,
    initAuth
};

// Initialize on DOM content loaded
document.addEventListener('DOMContentLoaded', initAuth);
