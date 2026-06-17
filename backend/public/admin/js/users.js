/* Users Index Page Script */

let currentPage = 1;
let totalPages = 1;
const limit = 10;
const rolesByID = new Map();

document.addEventListener('DOMContentLoaded', async function() {
    await loadRoles();
    setupFilters();
    setupLogout();
    await loadUsers();
});

function setupFilters() {
    const searchInput = document.getElementById('userSearch');
    const roleFilter = document.getElementById('roleFilter');
    const statusFilter = document.getElementById('statusFilter');
    const prevPage = document.getElementById('prevPage');
    const nextPage = document.getElementById('nextPage');

    let searchTimer;
    if (searchInput) {
        searchInput.addEventListener('input', () => {
            clearTimeout(searchTimer);
            searchTimer = setTimeout(() => {
                currentPage = 1;
                loadUsers();
            }, 250);
        });
    }

    [roleFilter, statusFilter].forEach((control) => {
        if (!control) return;
        control.addEventListener('change', () => {
            currentPage = 1;
            loadUsers();
        });
    });

    if (prevPage) {
        prevPage.addEventListener('click', () => {
            if (currentPage <= 1) return;
            currentPage--;
            loadUsers();
        });
    }

    if (nextPage) {
        nextPage.addEventListener('click', () => {
            if (currentPage >= totalPages) return;
            currentPage++;
            loadUsers();
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
    const roleFilter = document.getElementById('roleFilter');
    if (!roleFilter) return;

    try {
        const response = await fetch(`${API_BASE_URL}/admin/roles?limit=100`, {
            method: 'GET',
            headers: authRequestHeaders(),
            credentials: 'include'
        });
        if (!response.ok) throw new Error(`Failed to fetch roles: ${response.status}`);

        const payload = await response.json();
        const roles = Array.isArray(payload) ? payload : (payload.data?.roles || []);
        rolesByID.clear();
        roleFilter.innerHTML = '<option value="">All Roles</option>';

        roles.forEach((role) => {
            rolesByID.set(role.id, role.name);
            const option = document.createElement('option');
            option.value = role.id;
            option.textContent = role.name;
            roleFilter.appendChild(option);
        });
    } catch (error) {
        console.error('Failed to load roles:', error);
        showNotification('Failed to load roles', 'error');
    }
}

async function loadUsers() {
    const tableBody = document.querySelector('#users-table tbody');
    if (!tableBody) return;

    const searchVal = encodeURIComponent(document.getElementById('userSearch')?.value.trim() || '');
    const roleFilter = encodeURIComponent(document.getElementById('roleFilter')?.value || '');
    const statusFilter = encodeURIComponent(document.getElementById('statusFilter')?.value || '');
    const url = `${API_BASE_URL}/admin/users?page=${currentPage}&search=${searchVal}&role_id=${roleFilter}&is_active=${statusFilter}&limit=${limit}`;

    try {
        tableBody.innerHTML = '<tr><td colspan="8" class="text-center">Loading users...</td></tr>';

        const response = await fetch(url, {
            method: 'GET',
            headers: authRequestHeaders(),
            credentials: 'include'
        });
        if (!response.ok) throw new Error(`Failed to fetch users: ${response.status}`);

        const payload = await response.json();
        const users = payload.data?.users || [];
        const pagination = payload.data?.pagination || { page: 1, total_pages: 1, total: 0, limit };

        currentPage = pagination.page || 1;
        totalPages = pagination.total_pages || 1;

        renderUsers(tableBody, users);
        renderPagination(pagination);
    } catch (error) {
        console.error('Failed to load users:', error);
        tableBody.innerHTML = `<tr><td colspan="8" class="text-center">Error loading users: ${escapeHTML(error.message)}</td></tr>`;
        showNotification('Failed to load users', 'error');
    }
}

function renderUsers(tableBody, users) {
    tableBody.innerHTML = '';

    if (!users.length) {
        tableBody.innerHTML = '<tr><td colspan="8" class="text-center">No users found</td></tr>';
        return;
    }

    users.forEach((user, index) => {
        const serialNumber = ((currentPage - 1) * limit) + (index + 1);
        const tr = document.createElement('tr');
        tr.dataset.userId = user.id;

        let lastLogin = 'Never';
        if (user.last_login_at) {
            lastLogin = new Date(user.last_login_at).toLocaleString();
        }

        const roleName = rolesByID.get(user.role_id) || user.role?.name || '-';
        const fullName = `${user.first_name || ''} ${user.last_name || ''}`.trim();

        tr.innerHTML = `
            <td data-label="#">${serialNumber}</td>
            <td data-label="First Name">${escapeHTML(user.first_name || '')}</td>
            <td data-label="Last Name">${escapeHTML(user.last_name || '')}</td>
            <td data-label="Email">${escapeHTML(user.email || '')}</td>
            <td data-label="Role"><span class="role-badge">${escapeHTML(roleName)}</span></td>
            <td data-label="Status"><span class="status-badge status-${user.is_active ? 'active' : 'inactive'}">${user.is_active ? 'Active' : 'Inactive'}</span></td>
            <td data-label="Last Login">${escapeHTML(lastLogin)}</td>
            <td data-label="Actions" class="actions-col">
                <a href="/admin/users/edit.html?id=${encodeURIComponent(user.id || '')}" class="btn btn-outline btn-sm">Edit</a>
                <button class="btn btn-outline btn-sm btn-delete" data-user-id="${escapeHTML(user.id || '')}" data-user-name="${escapeHTML(fullName || user.email || 'User')}">Delete</button>
            </td>
        `;

        tableBody.appendChild(tr);
    });

    setupDeleteListeners();
}

function renderPagination(pagination) {
    const paginationInfo = document.getElementById('paginationInfo');
    const prevPage = document.getElementById('prevPage');
    const nextPage = document.getElementById('nextPage');

    if (paginationInfo) {
        paginationInfo.textContent = `Page ${pagination.page || 1} of ${pagination.total_pages || 1} (${pagination.total || 0} users)`;
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
            const userId = e.currentTarget.dataset.userId;
            const userName = e.currentTarget.dataset.userName || `User ${userId}`;

            if (!confirm(`Are you sure you want to delete the user "${userName}"? This action cannot be undone.`)) return;

            try {
                const response = await fetch(`${API_BASE_URL}/admin/users/${encodeURIComponent(userId)}`, {
                    method: 'DELETE',
                    headers: authRequestHeaders(),
                    credentials: 'include'
                });
                if (!response.ok) {
                    const errorData = await response.json().catch(() => ({}));
                    throw new Error(errorData.error || 'Failed to delete user');
                }

                showNotification('User deleted successfully', 'success');
                await loadUsers();
            } catch (error) {
                console.error('Failed to delete user:', error);
                showNotification(`Failed to delete user: ${error.message}`, 'error');
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
