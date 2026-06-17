/* Users Edit Page Script */

document.addEventListener('DOMContentLoaded', async function() {
    // Load roles for the dropdown
    await loadRoles();

    // Check if we're editing an existing user
    const urlParams = new URLSearchParams(window.location.search);
    const userId = urlParams.get('id');

    const form = document.getElementById('user-form');
    if (!form) return;

    // Set page title based on whether we're adding or editing
    const pageTitle = document.getElementById('page-title');
    if (pageTitle) {
        pageTitle.textContent = userId ? 'Edit User' : 'Add New User';
    }

    // If editing, load the user data
    if (userId) {
        await loadUserData(userId);
    }

    // Set up form submission
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        await saveUser(userId);
    });

    // Set up cancel button
    const cancelBtn = document.getElementById('cancel-btn');
    if (cancelBtn) {
        cancelBtn.addEventListener('click', (e) => {
            e.preventDefault();
            window.location.href = '/admin/users/';
        });
    }

    // Set up logout button
    const logoutBtn = document.getElementById('logout-btn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', async (e) => {
            e.preventDefault();
            const success = await auth.logout();
            if (!success) showNotification('Logout failed', 'error');
            window.location.href = 'http://localhost:5500/admin/login.html';
        });
    }
});

/**
 * Load all roles to populate the dropdown
 */
async function loadRoles() {
    try {
        const response = await api.get('/roles', { limit: 100 });
        if (!response.ok) {
            throw new Error(`Failed to fetch roles: ${response.status}`);
        }
        const payload = await response.json();
        const roles = Array.isArray(payload) ? payload : (payload.data?.roles || []);

        const roleSelects = document.querySelectorAll('#edit-role, #add-role, select[name="role_id"]');
        roleSelects.forEach(select => {
            // Clear existing options
            select.innerHTML = '<option value="">Select a role</option>';

            // Add role options
            roles.forEach(role => {
                const option = document.createElement('option');
                option.value = role.id;
                option.textContent = `${role.name} (${role.description || ''})`;
                select.appendChild(option);
            });
        });
    } catch (error) {
        console.error('Failed to load roles:', error);
        showNotification('Failed to load roles', 'error');
    }
}

/**
 * Load user data for editing
 * @param {string} userId - The ID of the user to load
 */
async function loadUserData(userId) {
    try {
        const response = await api.get(`/users/${userId}`);
        if (!response.ok) {
            throw new Error(`Failed to fetch user: ${response.status}`);
        }
        const user = await response.json();

        // Populate form fields
        document.getElementById('user-id').value = user.id || '';
        document.getElementById('first-name').value = user.first_name || '';
        document.getElementById('last-name').value = user.last_name || '';
        document.getElementById('email').value = user.email || '';
        document.getElementById('role').value = user.role_id || '';
        document.getElementById('is-active').value = user.is_active ? 'true' : 'false';
        // Password field remains empty for editing (unless user wants to change it)
    } catch (error) {
        console.error('Failed to load user data:', error);
        showNotification('Failed to load user data', 'error');
        // Redirect back to users list
        window.location.href = '/admin/users/';
    }
}

/**
 * Save the user (create or update)
 * @param {string|null} userId - The ID of the user if updating, null if creating
 */
async function saveUser(userId) {
    const form = document.getElementById('user-form');
    if (!form) return;

    // Gather form data
    const userData = {
        first_name: document.getElementById('first-name').value.trim(),
        last_name: document.getElementById('last-name').value.trim(),
        email: document.getElementById('email').value.trim(),
        role_id: document.getElementById('role').value,
        is_active: document.getElementById('is-active').value === 'true'
    };

    // Only include password if provided
    const password = document.getElementById('password').value;
    if (password) {
        userData.password = password;
    }

    // Validate required fields
    if (!userData.first_name || !userData.last_name || !userData.email || !userData.role_id) {
        showNotification('Please fill in all required fields', 'error');
        return;
    }

    try {
        const url = userId
            ? `${API_BASE_URL}/admin/users/${encodeURIComponent(userId)}`
            : `${API_BASE_URL}/admin/users`;
        const response = await fetch(url, {
            method: userId ? 'PUT' : 'POST',
            credentials: 'include',
            headers: authRequestHeaders(),
            body: JSON.stringify(userData)
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Failed to save user');
        }

        showNotification(userId ? 'User updated successfully' : 'User created successfully', 'success');
        // Redirect back to users list
        setTimeout(() => {
            window.location.href = '/admin/users/';
        }, 1500);
    } catch (error) {
        console.error('Failed to save user:', error);
        showNotification(`Failed to save user: ${error.message}`, 'error');
    }
}

function authRequestHeaders() {
    const headers = { 'Content-Type': 'application/json' };
    const token = localStorage.getItem('jwt') || sessionStorage.getItem('jwt');
    if (token) {
        headers.Authorization = `Bearer ${token}`;
    }
    return headers;
}
