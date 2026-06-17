/* Roles Index Page Script */

let currentPage = 1;
let totalPages = 1;
const limit = 10;

document.addEventListener('DOMContentLoaded', function() {
    setupFilters();
    setupLogout();
    loadRoles();
});

function setupFilters() {
    const searchInput = document.getElementById('roleSearch');
    const prevPage = document.getElementById('rolePrevPage');
    const nextPage = document.getElementById('roleNextPage');

    let searchTimer;
    if (searchInput) {
        searchInput.addEventListener('input', () => {
            clearTimeout(searchTimer);
            searchTimer = setTimeout(() => {
                currentPage = 1;
                loadRoles();
            }, 250);
        });
    }

    if (prevPage) {
        prevPage.addEventListener('click', () => {
            if (currentPage <= 1) return;
            currentPage--;
            loadRoles();
        });
    }

    if (nextPage) {
        nextPage.addEventListener('click', () => {
            if (currentPage >= totalPages) return;
            currentPage++;
            loadRoles();
        });
    }
}

function setupLogout() {
    const logoutBtn = document.getElementById('logout-btn');
    if (!logoutBtn) return;

    logoutBtn.addEventListener('click', async (e) => {
        e.preventDefault();
        const success = await auth.logout();
        if (!success) showNotification('Logout failed', 'error');
        window.location.href = 'http://localhost:5500/admin/login.html';
    });
}

async function loadRoles() {
    const tableBody = document.querySelector('#roles-table tbody');
    if (!tableBody) return;

    const searchVal = encodeURIComponent(document.getElementById('roleSearch')?.value.trim() || '');
    const url = `${API_BASE_URL}/admin/roles?page=${currentPage}&search=${searchVal}&limit=${limit}`;

    try {
        tableBody.innerHTML = '<tr><td colspan="6" class="text-center">Loading roles...</td></tr>';

        const response = await fetch(url, {
            method: 'GET',
            headers: authRequestHeaders(),
            credentials: 'include'
        });
        if (!response.ok) throw new Error(`Failed to fetch roles: ${response.status}`);

        const payload = await response.json();
        const roles = Array.isArray(payload) ? payload : (payload.data?.roles || []);
        const pagination = payload.data?.pagination || { page: 1, total_pages: 1, total: roles.length, limit };

        currentPage = pagination.page || 1;
        totalPages = pagination.total_pages || 1;

        renderRoles(tableBody, roles);
        renderPagination(pagination);
    } catch (error) {
        console.error('Failed to load roles:', error);
        tableBody.innerHTML = `<tr><td colspan="6" class="text-center">Error loading roles: ${escapeHTML(error.message)}</td></tr>`;
        showNotification('Failed to load roles', 'error');
    }
}

function renderRoles(tableBody, roles) {
    tableBody.innerHTML = '';

    if (!roles.length) {
        tableBody.innerHTML = '<tr><td colspan="6" class="text-center">No roles found</td></tr>';
        return;
    }

    roles.forEach((role, index) => {
        const serialNumber = ((currentPage - 1) * limit) + (index + 1);
        const tr = document.createElement('tr');
        tr.dataset.roleId = role.id;

        const permissionsHtml = role.permissions?.length
            ? role.permissions.map((permission) => `<span class="permission-tag">${escapeHTML(permission)}</span>`).join('')
            : '<span class="text-muted">No permissions</span>';

        const createdAt = role.created_at ? new Date(role.created_at).toLocaleString() : '-';

        tr.innerHTML = `
            <td data-label="#">${serialNumber}</td>
            <td data-label="Name">${escapeHTML(role.name || '')}</td>
            <td data-label="Description">${escapeHTML(role.description || '')}</td>
            <td data-label="Permissions">${permissionsHtml}</td>
            <td data-label="Created At">${escapeHTML(createdAt)}</td>
            <td data-label="Actions" class="actions-col">
                <a href="/admin/roles/edit.html?id=${encodeURIComponent(role.id || '')}" class="btn btn-outline btn-sm">Edit</a>
                <button class="btn btn-outline btn-sm btn-delete" data-role-id="${escapeHTML(role.id || '')}" data-role-name="${escapeHTML(role.name || 'Role')}">Delete</button>
            </td>
        `;

        tableBody.appendChild(tr);
    });

    setupDeleteListeners();
}

function renderPagination(pagination) {
    const paginationInfo = document.getElementById('rolePaginationInfo');
    const prevPage = document.getElementById('rolePrevPage');
    const nextPage = document.getElementById('roleNextPage');

    if (paginationInfo) {
        paginationInfo.textContent = `Page ${pagination.page || 1} of ${pagination.total_pages || 1} (${pagination.total || 0} roles)`;
    }
    if (prevPage) {
        prevPage.disabled = (pagination.page || 1) <= 1;
    }
    if (nextPage) {
        nextPage.disabled = (pagination.page || 1) >= (pagination.total_pages || 1);
    }
}

function setupDeleteListeners() {
    document.querySelectorAll('.btn-delete').forEach((button) => {
        button.addEventListener('click', async (e) => {
            const roleId = e.currentTarget.dataset.roleId;
            const roleName = e.currentTarget.dataset.roleName || `Role ${roleId}`;

            if (!confirm(`Are you sure you want to delete the role "${roleName}"? This action cannot be undone.`)) return;

            try {
                const response = await fetch(`${API_BASE_URL}/admin/roles/${encodeURIComponent(roleId)}`, {
                    method: 'DELETE',
                    headers: authRequestHeaders(),
                    credentials: 'include'
                });
                if (!response.ok) {
                    const errorData = await response.json().catch(() => ({}));
                    throw new Error(errorData.error || 'Failed to delete role');
                }

                showNotification('Role deleted successfully', 'success');
                await loadRoles();
            } catch (error) {
                console.error('Failed to delete role:', error);
                showNotification(`Failed to delete role: ${error.message}`, 'error');
            }
        });
    });
}

function escapeHTML(value) {
    return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

function authRequestHeaders() {
    const headers = {};
    const token = localStorage.getItem('jwt') || sessionStorage.getItem('jwt');
    if (token) {
        headers.Authorization = `Bearer ${token}`;
    }
    return headers;
}
