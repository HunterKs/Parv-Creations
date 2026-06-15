/* Roles Edit Page Script */

document.addEventListener('DOMContentLoaded', async function() {
    // Check if we're editing an existing role
    const urlParams = new URLSearchParams(window.location.search);
    const roleId = urlParams.get('id');

    const form = document.getElementById('role-form');
    if (!form) return;

    // Set page title based on whether we're adding or editing
    const pageTitle = document.getElementById('page-title');
    if (pageTitle) {
        pageTitle.textContent = roleId ? 'Edit Role' : 'Add New Role';
    }

    // If editing, load the role data
    if (roleId) {
        await loadRoleData(roleId);
    }

    // Set up form submission
    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        await saveRole(roleId);
    });

    // Set up cancel button
    const cancelBtn = document.getElementById('cancel-btn');
    if (cancelBtn) {
        cancelBtn.addEventListener('click', (e) => {
            e.preventDefault();
            window.location.href = '/admin/roles/';
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
 * Load role data for editing
 * @param {string} roleId - The ID of the role to load
 */
async function loadRoleData(roleId) {
    try {
        const response = await api.get(`/roles/${roleId}`);
        if (!response.ok) {
            throw new Error(`Failed to fetch role: ${response.status}`);
        }
        const role = await response.json();

        // Populate form fields
        document.getElementById('role-id').value = role.id || '';
        document.getElementById('name').value = role.name || '';
        document.getElementById('description').value = role.description || '';

        // Populate permissions checkboxes (we'll need to implement this based on how we store permissions)
        // For now, we'll just note that we need to implement permission loading
        // This would require fetching all available permissions or having a predefined list
        // Since we don't have a predefined list of permissions in the backend, we'll need to adjust
        // For simplicity, we'll assume we have a way to get all possible permissions
        // In a real app, we might have a permissions endpoint or a predefined list
        await loadPermissionsForRole(role.permissions || []);
    } catch (error) {
        console.error('Failed to load role data:', error);
        showNotification('Failed to load role data', 'error');
        // Redirect back to roles list
        window.location.href = '/admin/roles/';
    }
}

/**
 * Load permissions for a role (to populate checkboxes)
 * This is a simplified implementation - in a real app, we might have a predefined set of permissions
 * or fetch them from an endpoint
 * @param {Array<string>} selectedPermissions - Array of permission strings that should be checked
 */
async function loadPermissionsForRole(selectedPermissions = []) {
    // Since we don't have a predefined list of permissions in the backend,
    // we'll use a common set of permissions for demonstration
    // In a real app, you would fetch this from /api/v1/admin/permissions or similar
    const commonPermissions = [
        'users:read', 'users:create', 'users:update', 'users:delete',
        'roles:read', 'roles:create', 'roles:update', 'roles:delete',
        'products:read', 'products:create', 'products:update', 'products:delete',
        'orders:read', 'orders:create', 'orders:update', 'orders:delete'
    ];

    const container = document.querySelector('.permissions-container');
    if (!container) return;

    // Clear existing content
    container.innerHTML = '';

    // Create checkboxes for each permission
    commonPermissions.forEach(permission => {
        const label = document.createElement('label');
        label.className = 'permission-checkbox';
        label.style.display = 'flex';
        label.style.alignItems = 'center';
        label.style.margin = '0.25rem 0';

        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.value = permission;
        checkbox.checked = selectedPermissions.includes(permission);

        const span = document.createElement('span');
        span.textContent = permission;
        span.style.marginLeft = '0.5rem';

        label.appendChild(checkbox);
        label.appendChild(span);
        container.appendChild(label);
    });
}

/**
 * Save the role (create or update)
 * @param {string|null} roleId - The ID of the role if updating, null if creating
 */
async function saveRole(roleId) {
    const form = document.getElementById('role-form');
    if (!form) return;

    // Gather form data
    const roleData = {
        name: document.getElementById('name').value.trim(),
        description: document.getElementById('description').value.trim()
    };

    // Gather selected permissions
    const permissionCheckboxes = document.querySelectorAll('.permissions-container input[type="checkbox"]:checked');
    const permissions = Array.from(permissionCheckboxes).map(cb => cb.value);
    roleData.permissions = permissions;

    // Validate required fields
    if (!roleData.name) {
        showNotification('Please enter a role name', 'error');
        return;
    }

    if (roleData.permissions.length === 0) {
        showNotification('Please select at least one permission', 'error');
        return;
    }

    try {
        let response;
        if (roleId) {
            // Update existing role
            response = await api.put(`/roles/${roleId}`, roleData);
        } else {
            // Create new role
            response = await api.post('/roles', roleData);
        }

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Failed to save role');
        }

        showNotification(roleId ? 'Role updated successfully' : 'Role created successfully', 'success');
        // Redirect back to roles list
        setTimeout(() => {
            window.location.href = '/admin/roles/';
        }, 1500);
    } catch (error) {
        console.error('Failed to save role:', error);
        showNotification(`Failed to save role: ${error.message}`, 'error');
    }
}
