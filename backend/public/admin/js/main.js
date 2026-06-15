/* Main Dashboard Page Script */

document.addEventListener('DOMContentLoaded', function() {
    loadUserInfo();

    const logoutBtn = document.getElementById('logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', (e) => {
            e.preventDefault();

            fetch(API_BASE_URL + '/admin/auth/logout', {
                method: 'POST',
                credentials: 'include',
                headers: { 'Content-Type': 'application/json' }
            })
            .then(res => {
                if (res.ok) {
                    localStorage.removeItem('session_token');
                    sessionStorage.clear();
                    window.location.href = '/admin/login.html';
                } else {
                    showToast('Logout failed');
                }
            })
            .catch(err => {
                console.error(err);
                showToast('Logout failed');
            });
        });
    }
});

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

function showToast(message) {
    if (typeof showNotification === 'function') {
        showNotification(message, 'error');
        return;
    }
    console.error(message);
}
