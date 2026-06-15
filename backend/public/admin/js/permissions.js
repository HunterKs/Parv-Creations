/* Permissions Management Page Script */

let permissions = [];

document.addEventListener('DOMContentLoaded', function() {
    loadPermissions();

    const addButton = document.getElementById('add-permission-btn');
    const cancelButton = document.getElementById('cancel-permission-btn');
    const form = document.getElementById('permission-form');

    if (addButton) {
        addButton.addEventListener('click', () => openPermissionForm());
    }
    if (cancelButton) {
        cancelButton.addEventListener('click', () => closePermissionForm());
    }
    if (form) {
        form.addEventListener('submit', async (event) => {
            event.preventDefault();
            await savePermission();
        });
    }
});

async function loadPermissions() {
    const tableBody = document.querySelector('#permissions-table tbody');
    if (!tableBody) return;

    tableBody.innerHTML = '<tr><td colspan="5" class="text-center">Loading permissions...</td></tr>';

    try {
        const response = await fetch(`${API_BASE_URL}/permissions`, {
            method: 'GET',
            credentials: 'include'
        });
        if (!response.ok) throw new Error(`Failed to fetch permissions: ${response.status}`);

        permissions = await response.json();
        renderPermissions(tableBody);
    } catch (error) {
        console.error('Failed to load permissions:', error);
        tableBody.innerHTML = '<tr><td colspan="5" class="text-center">Failed to load permissions</td></tr>';
        showNotification(`Failed to load permissions: ${error.message}`, 'error');
    }
}

function renderPermissions(tableBody) {
    tableBody.innerHTML = '';

    if (!permissions.length) {
        tableBody.innerHTML = '<tr><td colspan="5" class="text-center">No permissions found</td></tr>';
        return;
    }

    permissions.forEach((permission) => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td data-label="ID">${escapeHTML(permission.id || '')}</td>
            <td data-label="Key"><span class="permission-tag">${escapeHTML(permission.key || '')}</span></td>
            <td data-label="Name">${escapeHTML(permission.name || '')}</td>
            <td data-label="Description">${escapeHTML(permission.description || '')}</td>
            <td data-label="Actions" class="actions-col">
                <div class="btn-group">
                    <button type="button" class="btn btn-sm btn-primary" data-action="edit" data-id="${escapeHTML(permission.id || '')}">Edit</button>
                    <button type="button" class="btn btn-sm btn-danger" data-action="delete" data-id="${escapeHTML(permission.id || '')}">Delete</button>
                </div>
            </td>
        `;
        tableBody.appendChild(row);
    });

    tableBody.querySelectorAll('[data-action="edit"]').forEach((button) => {
        button.addEventListener('click', () => {
            const permission = permissions.find((item) => item.id === button.dataset.id);
            openPermissionForm(permission);
        });
    });

    tableBody.querySelectorAll('[data-action="delete"]').forEach((button) => {
        button.addEventListener('click', () => deletePermission(button.dataset.id));
    });
}

function openPermissionForm(permission = null) {
    document.getElementById('permission-id').value = permission ? permission.id : '';
    document.getElementById('permission-key').value = permission ? permission.key : '';
    document.getElementById('permission-name').value = permission ? permission.name : '';
    document.getElementById('permission-description').value = permission ? permission.description || '' : '';
    document.getElementById('save-permission-btn').textContent = permission ? 'Update Permission' : 'Save Permission';
    document.getElementById('permission-form').hidden = false;
}

function closePermissionForm() {
    document.getElementById('permission-form').reset();
    document.getElementById('permission-id').value = '';
    document.getElementById('permission-form').hidden = true;
}

async function savePermission() {
    const id = document.getElementById('permission-id').value;
    const payload = {
        key: document.getElementById('permission-key').value.trim(),
        name: document.getElementById('permission-name').value.trim(),
        description: document.getElementById('permission-description').value.trim()
    };

    if (!payload.key || !payload.name) {
        showNotification('Permission key and name are required', 'error');
        return;
    }

    const url = id ? `${API_BASE_URL}/permissions/${id}` : `${API_BASE_URL}/permissions`;
    const method = id ? 'PUT' : 'POST';

    try {
        const response = await fetch(url, {
            method,
            credentials: 'include',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        if (!response.ok) {
            const error = await response.json().catch(() => ({}));
            throw new Error(error.error || `Permission save failed: ${response.status}`);
        }

        closePermissionForm();
        showNotification(id ? 'Permission updated' : 'Permission created', 'success');
        await loadPermissions();
    } catch (error) {
        console.error('Failed to save permission:', error);
        showNotification(`Failed to save permission: ${error.message}`, 'error');
    }
}

async function deletePermission(id) {
    if (!id || !confirm('Delete this permission?')) return;

    try {
        const response = await fetch(`${API_BASE_URL}/permissions/${id}`, {
            method: 'DELETE',
            credentials: 'include'
        });
        if (!response.ok) {
            const error = await response.json().catch(() => ({}));
            throw new Error(error.error || `Permission delete failed: ${response.status}`);
        }

        showNotification('Permission deleted', 'success');
        await loadPermissions();
    } catch (error) {
        console.error('Failed to delete permission:', error);
        showNotification(`Failed to delete permission: ${error.message}`, 'error');
    }
}

function escapeHTML(value) {
    return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}
