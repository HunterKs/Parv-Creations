/* Users Index Page Script */

document.addEventListener('DOMContentLoaded', function() {
    // Load users table
    loadUsers();

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
 * Load users from the API and populate the table
 */
async function loadUsers() {
    const tableBody = document.querySelector('#users-table tbody');
    if (!tableBody) return;

    try {
        // Show loading state
        tableBody.innerHTML = '<tr><td colspan="8" class="text-center">Loading users...</td></tr>';

        // Fetch users from API
        const response = await api.get('/users');
        if (!response.ok) {
            throw new Error(`Failed to fetch users: ${response.status}`);
        }

        const users = await response.json();

        // Clear table body
        tableBody.innerHTML = '';

        if (users.length === 0) {
            tableBody.innerHTML = '<tr><td colspan="8" class="text-center">No users found</td></tr>';
            return;
        }

        // Populate table rows
        users.forEach(user => {
            const tr = document.createElement('tr');
            tr.dataset.userId = user.id;

            // Format last login
            let lastLogin = 'Never';
            if (user.last_login_at) {
                const date = new Date(user.last_login_at);
                lastLogin = date.toLocaleString();
            }

            tr.innerHTML = `
                <td>${user.id}</td>
                <td>${user.first_name || ''}</td>
                <td>${user.last_name || ''}</td>
                <td>${user.email || ''}</td>
                <td><span class="role-badge">${user.role_name || '-'}</span></td>
                <td><span class="status-badge status-${user.is_active ? 'active' : 'inactive'}">${user.is_active ? 'Active' : 'Inactive'}</span></td>
                <td>${lastLogin}</td>
                <td class="actions-col">
                    <a href="/admin/users/edit.html?id=${user.id}" class="btn btn-outline btn-sm">Edit</a>
                    <button class="btn btn-outline btn-sm btn-delete" data-user-id="${user.id}" data-user-name="${user.first_name} ${user.last_name}">Delete</button>
                </td>
            `;

            tableBody.appendChild(tr);
        });

        // Set up delete button event listeners
        setupDeleteListeners();
    } catch (error) {
        console.error('Failed to load users:', error);
        tableBody.innerHTML = `<tr><td colspan="8" class="text-center">Error loading users: ${error.message}</td></tr>`;
        showNotification('Failed to load users', 'error');
    }
}

/**
 * Set up event listeners for delete buttons
 */
function setupDeleteListeners() {
    const deleteButtons = document.querySelectorAll('.btn-delete');
    deleteButtons.forEach(button => {
        button.addEventListener('click', async (e) => {
            const userId = e.target.dataset.userId;
            const userName = e.target.dataset.userName || `User ${userId}`;

            if (confirm(`Are you sure you want to delete the user "${userName}"? This action cannot be undone.`)) {
                try {
                    const response = await api.delete(`/users/${userId}`);
                    if (response.ok) {
                        showNotification('User deleted successfully', 'success');
                        // Remove the row from the table
                        e.target.closest('tr').remove();
                    } else {
                        const errorData = await response.json();
                        throw new Error(errorData.error || 'Failed to delete user');
                    }
                } catch (error) {
                    console.error('Failed to delete user:', error);
                    showNotification(`Failed to delete user: ${error.message}`, 'error');
                }
            }
        });
    });
}
