/* Roles Index Page Script */

document.addEventListener('DOMContentLoaded', function() {
    // Load roles table
    loadRoles();

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
    }
});

/**
 * Load roles from the API and populate the table
 */
async function loadRoles() {
    const tableBody = document.querySelector('#roles-table tbody');
    if (!tableBody) return;

    try {
        // Show loading state
        tableBody.innerHTML = '<tr><td colspan="6" class="text-center">Loading roles...</td></tr>';

        // Fetch roles from API
        const response = await api.get('/roles');
        if (!response.ok) {
            throw new Error(`Failed to fetch roles: ${response.status}`);
        }

        const roles = await response.json();

        // Clear table body
        tableBody.innerHTML = '';

        if (roles.length === 0) {
            tableBody.innerHTML = '<tr><td colspan="6" class="text-center">No roles found</td></tr>';
            return;
        }

        // Populate table rows
        roles.forEach(role => {
            const tr = document.createElement('tr');
            tr.dataset.roleId = role.id;

            // Format permissions as tags
            const permissionsHtml = role.permissions
                ? role.permissions.map(p => `<span class="permission-tag">${p}</span>`).join('')
                : '<span class="text-muted">No permissions</span>';

            tr.innerHTML = `
                <td>${role.id}</td>
                <td>${role.name || ''}</td>
                <td>${role.description || ''}</td>
                <td>${permissionsHtml}</td>
                <td>${new Date(role.created_at).toLocaleString()}</td>
                <td class="actions-col">
                    <a href="/admin/roles/edit.html?id=${role.id}" class="btn btn-outline btn-sm">Edit</a>
                    <button class="btn btn-outline btn-sm btn-delete" data-role-id="${role.id}" data-role-name="${role.name}">Delete</button>
                </td>
            `;

            tableBody.appendChild(tr);
        });

        // Set up delete button event listeners
        setupDeleteListeners();
    } catch (error) {
        console.error('Failed to load roles:', error);
        tableBody.innerHTML = `<tr><td colspan="6" class="text-center">Error loading roles: ${error.message}</td></tr>`;
        showNotification('Failed to load roles', 'error');
    }
}

/**
 * Set up event listeners for delete buttons
 */
function setupDeleteListeners() {
    const deleteButtons = document.querySelectorAll('.btn-delete');
    deleteButtons.forEach(button => {
        button.addEventListener('click', async (e) => {
            const roleId = e.target.dataset.roleId;
            const roleName = e.target.dataset.roleName || `Role ${roleId}`;

            if (confirm(`Are you sure you want to delete the role "${roleName}"? This action cannot be undone.`)) {
                try {
                    const response = await api.delete(`/roles/${roleId}`);
                    if (response.ok) {
                        showNotification('Role deleted successfully', 'success');
                        // Remove the row from the table
                        e.target.closest('tr').remove();
                    } else {
                        const errorData = await response.json();
                        throw new Error(errorData.error || 'Failed to delete role');
                    }
                } catch (error) {
                    console.error('Failed to delete role:', error);
                    showNotification(`Failed to delete role: ${error.message}`, 'error');
                }
            }
        });
    });
}