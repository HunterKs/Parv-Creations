/* Main Dashboard Page Script */

document.addEventListener('DOMContentLoaded', function() {
    // Load username from auth if available
    loadUserInfo();

    // Set up logout button
    const logoutBtn = document.getElementById('logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', async (e) => {
            e.preventDefault();
            const success = await auth.logout();
            if (success) {
                window.location.href = '/admin/login.html';
            } else {
                showNotification('Logout failed', 'error');
            }
        });
    });
});

/**
 * Load and display the current user's information
 */
async function loadUserInfo() {
    try {
        const user = await auth.getCurrentUser();
        const usernameElement = document.getElementById('username');
        if (usernameElement && user) {
            usernameElement.textContent = `${user.first_name} ${user.last_name}`;
        }
    } catch (error) {
        console.error('Failed to load user info:', error);
    }
}