// API Configuration
const API_BASE = '/api';

// State Management
const state = {
    currentView: 'computes',
    theme: localStorage.getItem('theme') || 'light',
    data: {
        computes: [],
        components: [],
        services: [],
        assignments: [],
        ips: [],
        dns: [],
        firewall: [],
        ports: [],
        journal: [],
        apikeys: []
    }
};

// Initialize App
document.addEventListener('DOMContentLoaded', () => {
    initTheme();
    initNavigation();
    loadView('computes');
});

// Theme Management
function initTheme() {
    document.documentElement.setAttribute('data-theme', state.theme);
    const themeToggle = document.getElementById('theme-toggle');
    themeToggle.addEventListener('click', toggleTheme);
}

function toggleTheme() {
    state.theme = state.theme === 'light' ? 'dark' : 'light';
    document.documentElement.setAttribute('data-theme', state.theme);
    localStorage.setItem('theme', state.theme);
}

// Navigation
function initNavigation() {
    document.querySelectorAll('.nav-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const view = e.target.dataset.view;
            setActiveNav(view);
            loadView(view);
        });
    });
}

function setActiveNav(view) {
    document.querySelectorAll('.nav-btn').forEach(btn => {
        btn.classList.remove('active');
        if (btn.dataset.view === view) {
            btn.classList.add('active');
        }
    });
    state.currentView = view;
}

// View Loader
async function loadView(view) {
    const main = document.getElementById('main-content');
    main.innerHTML = '<div class="loading">Loading...</div>';

    try {
        switch(view) {
            case 'computes':
                await renderComputesView();
                break;
            case 'components':
                await renderComponentsView();
                break;
            case 'services':
                await renderServicesView();
                break;
            case 'assignments':
                await renderAssignmentsView();
                break;
            case 'ips':
                await renderIPsView();
                break;
            case 'dns':
                await renderDNSView();
                break;
            case 'firewall':
                await renderFirewallView();
                break;
            case 'ports':
                await renderPortsView();
                break;
            case 'journal':
                await renderJournalView();
                break;
            case 'apikeys':
                await renderAPIKeysView();
                break;
            case 'reports':
                await renderReportsView();
                break;
        }
    } catch (error) {
        main.innerHTML = `<div class="error">Error loading ${view}: ${error.message}</div>`;
    }
}

// Computes View
async function renderComputesView() {
    const computes = await fetchAPI('/computes');
    state.data.computes = computes;

    const html = `
        <div class="view-header">
            <h2>Compute Resources</h2>
            <button class="btn btn-primary" onclick="showCreateComputeModal()">+ Add Compute</button>
        </div>
        ${computes.length === 0 ? '<div class="empty-state"><p>No computes found</p></div>' : `
        <div class="table-wrapper">
        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Type</th>
                    <th>Provider</th>
                    <th>Region</th>
                    <th>State</th>
                    <th>Tags</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${computes.map(c => `
                    <tr>
                        <td><strong>${escapeHtml(c.name)}</strong></td>
                        <td>${escapeHtml(c.type)}</td>
                        <td>${escapeHtml(c.provider)}</td>
                        <td>${escapeHtml(c.region)}</td>
                        <td class="status-${c.state}">${escapeHtml(c.state)}</td>
                        <td>${renderTags(c.tags)}</td>
                        <td class="actions">
                            <button class="btn btn-danger" onclick="deleteCompute('${c.id}')">Delete</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
        </div>
        `}
    `;

    document.getElementById('main-content').innerHTML = html;
}

// Components View
async function renderComponentsView() {
    const components = await fetchAPI('/components');
    state.data.components = components;

    const html = `
        <div class="view-header">
            <h2>Hardware Components</h2>
            <button class="btn btn-primary" onclick="showCreateComponentModal()">+ Add Component</button>
        </div>
        ${components.length === 0 ? '<div class="empty-state"><p>No components found</p></div>' : `
        <div class="table-wrapper">
        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Type</th>
                    <th>Manufacturer</th>
                    <th>Model</th>
                    <th>Specs</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${components.map(c => `
                    <tr>
                        <td><strong>${escapeHtml(c.name)}</strong></td>
                        <td>${escapeHtml(c.type)}</td>
                        <td>${escapeHtml(c.manufacturer)}</td>
                        <td>${escapeHtml(c.model)}</td>
                        <td class="wrap"><small>${JSON.stringify(c.specs)}</small></td>
                        <td class="actions">
                            <button class="btn btn-danger" onclick="deleteComponent('${c.id}')">Delete</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
        </div>
        `}
    `;

    document.getElementById('main-content').innerHTML = html;
}

// Services View
async function renderServicesView() {
    const services = await fetchAPI('/services');
    state.data.services = services;

    const html = `
        <div class="view-header">
            <h2>Services</h2>
            <button class="btn btn-primary" onclick="showCreateServiceModal()">+ Add Service</button>
        </div>
        ${services.length === 0 ? '<div class="empty-state"><p>No services found</p></div>' : `
        <div class="table-wrapper">
        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Min Resources</th>
                    <th>Max Resources</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${services.map(s => `
                    <tr>
                        <td><strong>${escapeHtml(s.name)}</strong></td>
                        <td class="wrap"><small>${JSON.stringify(s.min_spec)}</small></td>
                        <td class="wrap"><small>${JSON.stringify(s.max_spec)}</small></td>
                        <td class="actions">
                            <button class="btn btn-danger" onclick="deleteService('${s.id}')">Delete</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
        </div>
        `}
    `;

    document.getElementById('main-content').innerHTML = html;
}

// Assignments View
async function renderAssignmentsView() {
    const assignments = await fetchAPI('/assignments');
    state.data.assignments = assignments;

    // Fetch computes and services to resolve names
    if (state.data.computes.length === 0) {
        state.data.computes = await fetchAPI('/computes');
    }
    if (state.data.services.length === 0) {
        state.data.services = await fetchAPI('/services');
    }

    // Create lookup maps
    const computeMap = {};
    state.data.computes.forEach(c => computeMap[c.id] = c.name);
    const serviceMap = {};
    state.data.services.forEach(s => serviceMap[s.id] = s.name);

    const html = `
        <div class="view-header">
            <h2>Service Assignments</h2>
        </div>
        ${assignments.length === 0 ? '<div class="empty-state"><p>No assignments found</p></div>' : `
        <div class="table-wrapper">
        <table>
            <thead>
                <tr>
                    <th>Service</th>
                    <th>Compute</th>
                    <th>Allocated Resources</th>
                    <th>Created</th>
                </tr>
            </thead>
            <tbody>
                ${assignments.map(a => `
                    <tr>
                        <td><strong>${escapeHtml(serviceMap[a.service_id] || a.service_id)}</strong></td>
                        <td><strong>${escapeHtml(computeMap[a.compute_id] || a.compute_id)}</strong></td>
                        <td class="wrap"><small>${JSON.stringify(a.allocated)}</small></td>
                        <td>${new Date(a.created_at).toLocaleString()}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
        </div>
        `}
    `;

    document.getElementById('main-content').innerHTML = html;
}

// IPs View
async function renderIPsView() {
    const ips = await fetchAPI('/ips');
    state.data.ips = ips;

    const html = `
        <div class="view-header">
            <h2>IP Addresses</h2>
            <button class="btn btn-primary" onclick="showCreateIPModal()">+ Add IP</button>
        </div>
        ${ips.length === 0 ? '<div class="empty-state"><p>No IPs found</p></div>' : `
        <div class="table-wrapper">
        <table>
            <thead>
                <tr>
                    <th>Address</th>
                    <th>Type</th>
                    <th>CIDR</th>
                    <th>Gateway</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${ips.map(ip => `
                    <tr>
                        <td><strong>${escapeHtml(ip.address)}</strong></td>
                        <td>${escapeHtml(ip.type)}</td>
                        <td>${escapeHtml(ip.cidr || '-')}</td>
                        <td>${escapeHtml(ip.gateway || '-')}</td>
                        <td class="actions">
                            <button class="btn btn-danger" onclick="deleteIP('${ip.id}')">Delete</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
        </div>
        `}
    `;

    document.getElementById('main-content').innerHTML = html;
}

// DNS View
async function renderDNSView() {
    const dns = await fetchAPI('/dns');
    state.data.dns = dns;

    const html = `
        <div class="view-header">
            <h2>DNS Records</h2>
            <button class="btn btn-primary" onclick="showCreateDNSModal()">+ Add DNS Record</button>
        </div>
        ${dns.length === 0 ? '<div class="empty-state"><p>No DNS records found</p></div>' : `
        <div class="table-wrapper">
        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Type</th>
                    <th>Value</th>
                    <th>Zone</th>
                    <th>TTL</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${dns.map(d => `
                    <tr>
                        <td><strong>${escapeHtml(d.name)}</strong></td>
                        <td>${escapeHtml(d.type)}</td>
                        <td>${escapeHtml(d.value)}</td>
                        <td>${escapeHtml(d.zone)}</td>
                        <td>${d.ttl}</td>
                        <td class="actions">
                            <button class="btn btn-danger" onclick="deleteDNS('${d.id}')">Delete</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
        </div>
        `}
    `;

    document.getElementById('main-content').innerHTML = html;
}

// Firewall View
async function renderFirewallView() {
    const firewall = await fetchAPI('/firewall');
    state.data.firewall = firewall;

    const html = `
        <div class="view-header">
            <h2>Firewall Rules</h2>
            <button class="btn btn-primary" onclick="showCreateFirewallModal()">+ Add Rule</button>
        </div>
        ${firewall.length === 0 ? '<div class="empty-state"><p>No firewall rules found</p></div>' : `
        <div class="table-wrapper">
        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Action</th>
                    <th>Protocol</th>
                    <th>Source</th>
                    <th>Destination</th>
                    <th>Ports</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${firewall.map(f => `
                    <tr>
                        <td><strong>${escapeHtml(f.name)}</strong></td>
                        <td>${escapeHtml(f.action)}</td>
                        <td>${escapeHtml(f.protocol)}</td>
                        <td>${escapeHtml(f.source)}</td>
                        <td>${escapeHtml(f.destination)}</td>
                        <td>${f.port_start}${f.port_end !== f.port_start ? '-' + f.port_end : ''}</td>
                        <td class="actions">
                            <button class="btn btn-danger" onclick="deleteFirewall('${f.id}')">Delete</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
        </div>
        `}
    `;

    document.getElementById('main-content').innerHTML = html;
}

// Ports View
async function renderPortsView() {
    const ports = await fetchAPI('/ports');
    state.data.ports = ports;

    const html = `
        <div class="view-header">
            <h2>Port Mappings</h2>
            <button class="btn btn-primary" onclick="showCreatePortModal()">+ Add Port</button>
        </div>
        ${ports.length === 0 ? '<div class="empty-state"><p>No port mappings found</p></div>' : `
        <div class="table-wrapper">
        <table>
            <thead>
                <tr>
                    <th>Assignment ID</th>
                    <th>IP ID</th>
                    <th>External Port</th>
                    <th>Protocol</th>
                    <th>Service Port</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${ports.map(p => `
                    <tr>
                        <td>${escapeHtml(p.assignment_id)}</td>
                        <td>${escapeHtml(p.ip_id)}</td>
                        <td><strong>${p.port}</strong></td>
                        <td>${escapeHtml(p.protocol)}</td>
                        <td>${p.service_port}</td>
                        <td class="actions">
                            <button class="btn btn-danger" onclick="deletePort('${p.id}')">Delete</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
        </div>
        `}
    `;

    document.getElementById('main-content').innerHTML = html;
}

// Journal View
async function renderJournalView() {
    const journal = await fetchAPI('/journal');
    state.data.journal = journal;

    // Fetch computes to resolve names
    if (state.data.computes.length === 0) {
        state.data.computes = await fetchAPI('/computes');
    }

    const computeMap = {};
    state.data.computes.forEach(c => computeMap[c.id] = c.name);

    const html = `
        <div class="view-header">
            <h2>Journal Entries</h2>
            <button class="btn btn-primary" onclick="showCreateJournalModal()">+ Add Entry</button>
        </div>
        ${journal.length === 0 ? '<div class="empty-state"><p>No journal entries found</p></div>' : `
        <div class="table-wrapper">
        <table>
            <thead>
                <tr>
                    <th>Compute</th>
                    <th>Category</th>
                    <th>Content</th>
                    <th>Timestamp</th>
                </tr>
            </thead>
            <tbody>
                ${journal.map(j => `
                    <tr>
                        <td><strong>${escapeHtml(computeMap[j.compute_id] || j.compute_id)}</strong></td>
                        <td>${escapeHtml(j.category)}</td>
                        <td class="wrap">${escapeHtml(j.content)}</td>
                        <td>${new Date(j.timestamp).toLocaleString()}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
        </div>
        `}
    `;

    document.getElementById('main-content').innerHTML = html;
}

// API Keys View
async function renderAPIKeysView() {
    const apikeys = await fetchAPI('/admin/apikeys');
    state.data.apikeys = apikeys;

    const html = `
        <div class="view-header">
            <h2>API Keys</h2>
            <button class="btn btn-primary" onclick="showCreateAPIKeyModal()">+ Add API Key</button>
        </div>
        ${apikeys.length === 0 ? '<div class="empty-state"><p>No API keys found</p></div>' : `
        <div class="table-wrapper">
        <table>
            <thead>
                <tr>
                    <th>Name</th>
                    <th>Scope</th>
                    <th>Description</th>
                    <th>Created</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${apikeys.map(k => `
                    <tr>
                        <td><strong>${escapeHtml(k.name)}</strong></td>
                        <td>${escapeHtml(k.scope)}</td>
                        <td class="wrap">${escapeHtml(k.description || '-')}</td>
                        <td>${new Date(k.created_at).toLocaleString()}</td>
                        <td class="actions">
                            <button class="btn btn-danger" onclick="deleteAPIKey('${k.id}')">Delete</button>
                        </td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
        </div>
        `}
    `;

    document.getElementById('main-content').innerHTML = html;
}

// Reports View
async function renderReportsView() {
    // Fetch computes for dropdown
    if (state.data.computes.length === 0) {
        state.data.computes = await fetchAPI('/computes');
    }

    const html = `
        <div class="view-header">
            <h2>Compute Reports</h2>
        </div>
        <div class="report-container">
            <div class="select-compute">
                <label>Select Compute:</label>
                <select id="compute-select" onchange="loadComputeReport(this.value)">
                    <option value="">-- Select a compute --</option>
                    ${state.data.computes.map(c => `
                        <option value="${c.id}">${escapeHtml(c.name)} (${escapeHtml(c.provider)} - ${escapeHtml(c.region)})</option>
                    `).join('')}
                </select>
            </div>
            <div id="report-content"></div>
        </div>
    `;

    document.getElementById('main-content').innerHTML = html;
}

async function loadComputeReport(computeId) {
    if (!computeId) {
        document.getElementById('report-content').innerHTML = '';
        return;
    }

    const reportContent = document.getElementById('report-content');
    reportContent.innerHTML = '<div class="loading">Loading report...</div>';

    try {
        const report = await fetchAPI(`/reports/compute/${computeId}`);

        // Fetch component details for component assignments
        const componentMap = {};
        if (state.data.components.length === 0) {
            state.data.components = await fetchAPI('/components');
        }
        state.data.components.forEach(c => componentMap[c.id] = c);

        // Fetch service details for service assignments
        const serviceMap = {};
        if (state.data.services.length === 0) {
            state.data.services = await fetchAPI('/services');
        }
        state.data.services.forEach(s => serviceMap[s.id] = s);

        // Fetch IP details for IP assignments
        const ipMap = {};
        for (const ipAssignment of report.ip_assignments || []) {
            if (!ipMap[ipAssignment.ip_id]) {
                try {
                    const ip = await fetchAPI(`/ips/${ipAssignment.ip_id}`);
                    ipMap[ipAssignment.ip_id] = ip;
                } catch (e) {
                    // Ignore if IP not found
                }
            }
        }

        const compute = report.compute;

        // Calculate resource totals
        let totalCores = 0;
        let totalMemoryGB = 0;
        let totalVRAMGB = 0;
        let totalStorageGB = 0;
        const raidGroups = {};
        const nonRaidStorage = [];

        (report.component_assignments || []).forEach(ca => {
            const comp = componentMap[ca.component_id];
            if (!comp || !comp.specs) return;

            const getSpecFloat = (specs, ...keys) => {
                for (const key of keys) {
                    const val = specs[key];
                    if (typeof val === 'number') return val;
                }
                return 0;
            };

            switch (comp.type) {
                case 'cpu':
                    const threads = getSpecFloat(comp.specs, 'threads', 'thread_count', 'cores', 'core_count');
                    totalCores += threads * ca.quantity;
                    break;
                case 'ram':
                case 'memory':
                    let mem = getSpecFloat(comp.specs, 'capacity_gb', 'size', 'size_gb', 'memory_gb', 'memory');
                    if (mem > 1024) mem /= 1024;  // Convert MB to GB
                    totalMemoryGB += mem * ca.quantity;
                    break;
                case 'gpu':
                    let vram = getSpecFloat(comp.specs, 'vram_gb', 'vram', 'memory_gb', 'video_memory_gb', 'memory');
                    if (vram > 1024) vram /= 1024;  // Convert MB to GB
                    totalVRAMGB += vram * ca.quantity;
                    break;
                case 'storage':
                case 'nvme':
                case 'ssd':
                case 'hdd':
                    const storage = getSpecFloat(comp.specs, 'size', 'capacity_gb', 'storage_gb', 'capacity');
                    if (storage > 0) {
                        const si = {
                            size: storage,
                            quantity: ca.quantity,
                            raidLevel: ca.raid_level || '',
                            name: comp.name
                        };

                        if (ca.raid_level && ca.raid_level !== 'none' && ca.raid_group) {
                            if (!raidGroups[ca.raid_group]) {
                                raidGroups[ca.raid_group] = [];
                            }
                            raidGroups[ca.raid_group].push(si);
                        } else {
                            nonRaidStorage.push(si);
                        }
                    }
                    break;
            }
        });

        // Calculate RAID capacity
        const calculateRaidCapacity = (storage) => {
            if (storage.length === 0) return 0;
            const raidLevel = storage[0].raidLevel;
            const disks = [];
            storage.forEach(si => {
                for (let i = 0; i < si.quantity; i++) {
                    disks.push(si.size);
                }
            });

            if (disks.length === 0) return 0;

            switch (raidLevel) {
                case 'raid0':
                    return disks.reduce((a, b) => a + b, 0);
                case 'raid1':
                    return Math.min(...disks);
                case 'raid5':
                    return disks.length >= 3 ? (disks.length - 1) * Math.min(...disks) : disks.reduce((a, b) => a + b, 0);
                case 'raid6':
                    return disks.length >= 4 ? (disks.length - 2) * Math.min(...disks) : disks.reduce((a, b) => a + b, 0);
                case 'raid10':
                    return (disks.length >= 4 && disks.length % 2 === 0) ? disks.reduce((a, b) => a + b, 0) / 2 : disks.reduce((a, b) => a + b, 0);
                default:
                    return disks.reduce((a, b) => a + b, 0);
            }
        };

        Object.values(raidGroups).forEach(group => {
            totalStorageGB += calculateRaidCapacity(group);
        });
        nonRaidStorage.forEach(si => {
            totalStorageGB += si.size * si.quantity;
        });

        // Calculate allocated resources
        let allocatedCores = 0;
        let allocatedMemoryMB = 0;
        let allocatedVRAMMB = 0;
        let allocatedStorageGB = 0;

        (report.service_assignments || []).forEach(a => {
            if (a.allocated) {
                allocatedCores += a.allocated.cores || 0;
                allocatedMemoryMB += a.allocated.memory || 0;
                allocatedVRAMMB += a.allocated.vram || 0;
                allocatedStorageGB += (a.allocated.nvme || 0) + (a.allocated.sata || 0);
            }
        });
        // Calculate utilization percentages
        const totalMemoryMB = totalMemoryGB * 1024;
        const totalVRAMMB = totalVRAMGB * 1024;
        const coresUtil = totalCores > 0 ? (allocatedCores / totalCores * 100).toFixed(1) : '0.0';
        const memUtil = totalMemoryMB > 0 ? (allocatedMemoryMB / totalMemoryMB * 100).toFixed(1) : '0.0';
        const vramUtil = totalVRAMMB > 0 ? (allocatedVRAMMB / totalVRAMMB * 100).toFixed(1) : '0.0';
        const storageUtil = totalStorageGB > 0 ? (allocatedStorageGB / totalStorageGB * 100).toFixed(1) : '0.0';

        const html = `
            <div class="report-section">
                <h3>Compute Information</h3>
                <div class="report-info">
                    <div class="report-info-item">
                        <label>Name</label>
                        <div class="value"><strong>${escapeHtml(compute.name)}</strong></div>
                    </div>
                    <div class="report-info-item">
                        <label>Type</label>
                        <div class="value">${escapeHtml(compute.type)}</div>
                    </div>
                    <div class="report-info-item">
                        <label>Provider</label>
                        <div class="value">${escapeHtml(compute.provider)}</div>
                    </div>
                    <div class="report-info-item">
                        <label>Region</label>
                        <div class="value">${escapeHtml(compute.region)}</div>
                    </div>
                    <div class="report-info-item">
                        <label>State</label>
                        <div class="value status-${compute.state}">${escapeHtml(compute.state)}</div>
                    </div>
                    ${compute.tags && Object.keys(compute.tags).length > 0 ? `
                    <div class="report-info-item" style="grid-column: 1 / -1;">
                        <label>Tags</label>
                        <div class="value">${renderTags(compute.tags)}</div>
                    </div>
                    ` : ''}
                </div>
            </div>

            ${totalCores > 0 || totalMemoryGB > 0 || totalStorageGB > 0 ? `
            <div class="report-section">
                <h3>Resource Summary</h3>
                <div class="report-info">
                    ${totalCores > 0 ? `
                    <div class="report-info-item">
                        <label>Cores</label>
                        <div class="value">${totalCores} <small>(${coresUtil}% allocated)</small></div>
                    </div>` : ''}
                    ${totalMemoryGB > 0 ? `
                    <div class="report-info-item">
                        <label>Memory</label>
                        <div class="value">${Math.round(totalMemoryGB)} GB <small>(${memUtil}% allocated)</small></div>
                    </div>` : ''}
                    ${totalVRAMGB > 0 ? `
                    <div class="report-info-item">
                        <label>VRAM</label>
                        <div class="value">${Math.round(totalVRAMGB)} GB <small>(${vramUtil}% allocated)</small></div>
                    </div>` : ''}
                    ${totalStorageGB > 0 ? `
                    <div class="report-info-item">
                        <label>Storage</label>
                        <div class="value">${Math.round(totalStorageGB)} GB <small>(${storageUtil}% allocated)</small></div>
                    </div>` : ''}
                </div>

                ${Object.keys(raidGroups).length > 0 || nonRaidStorage.length > 0 ? `
                <h4 style="margin-top: 20px;">Storage Configuration</h4>
                <div class="table-wrapper">
                <table>
                    <thead>
                        <tr>
                            <th>Configuration</th>
                            <th>Components</th>
                            <th>Capacity</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${Object.entries(raidGroups).map(([groupId, group]) => {
                            const capacity = calculateRaidCapacity(group);
                            const raidLevel = group[0].raidLevel;
                            const diskCount = group.reduce((sum, si) => sum + si.quantity, 0);
                            const components = group.map(si => `${si.quantity}x ${si.name} (${Math.round(si.size)} GB each)`).join(', ');
                            return `
                            <tr>
                                <td><strong>RAID Group: ${escapeHtml(groupId)}</strong><br><small>${raidLevel.toUpperCase()} (${diskCount} disks)</small></td>
                                <td class="wrap"><small>${escapeHtml(components)}</small></td>
                                <td>${Math.round(capacity)} GB</td>
                            </tr>`;
                        }).join('')}
                        ${nonRaidStorage.length > 0 ? `
                        <tr>
                            <td><strong>Non-RAID Storage</strong></td>
                            <td class="wrap"><small>${nonRaidStorage.map(si => `${si.quantity}x ${si.name}`).join(', ')}</small></td>
                            <td>${Math.round(nonRaidStorage.reduce((sum, si) => sum + (si.size * si.quantity), 0))} GB</td>
                        </tr>` : ''}
                    </tbody>
                </table>
                </div>
                ` : ''}
            </div>
            ` : ''}

            ${report.component_assignments && report.component_assignments.length > 0 ? `
            <div class="report-section">
                <h3>Hardware Components</h3>
                <div class="table-wrapper">
                <table>
                    <thead>
                        <tr>
                            <th>Component</th>
                            <th>Type</th>
                            <th>Manufacturer</th>
                            <th>Model</th>
                            <th>Specs</th>
                            <th>Quantity</th>
                            <th>RAID</th>
                            <th>Notes</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${report.component_assignments.map(ca => {
                            const component = componentMap[ca.component_id] || {};
                            const specsHtml = component.specs && Object.keys(component.specs).length > 0
                                ? '<small>' + Object.entries(component.specs).map(([k, v]) => `${escapeHtml(k)}: ${escapeHtml(String(v))}`).join('<br>') + '</small>'
                                : '-';
                            return `
                            <tr>
                                <td><strong>${escapeHtml(component.name || ca.component_id)}</strong></td>
                                <td>${escapeHtml(component.type || '-')}</td>
                                <td>${escapeHtml(component.manufacturer || '-')}</td>
                                <td>${escapeHtml(component.model || '-')}</td>
                                <td class="wrap">${specsHtml}</td>
                                <td>${ca.quantity}</td>
                                <td>${ca.raid_level ? escapeHtml(ca.raid_level) + (ca.raid_group ? ' (' + escapeHtml(ca.raid_group) + ')' : '') : '-'}</td>
                                <td class="wrap">${escapeHtml(ca.notes || '-')}</td>
                            </tr>
                            `;
                        }).join('')}
                    </tbody>
                </table>
                </div>
            </div>
            ` : '<div class="report-section"><p>No hardware components assigned</p></div>'}

            ${report.service_assignments && report.service_assignments.length > 0 ? `
            <div class="report-section">
                <h3>Service Assignments</h3>
                <div class="table-wrapper">
                <table>
                    <thead>
                        <tr>
                            <th>Service</th>
                            <th>Allocated Resources</th>
                            <th>Created</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${report.service_assignments.map(a => {
                            const service = serviceMap[a.service_id] || {};
                            return `
                            <tr>
                                <td><strong>${escapeHtml(service.name || a.service_id)}</strong></td>
                                <td class="wrap"><small>${JSON.stringify(a.allocated)}</small></td>
                                <td>${new Date(a.created_at).toLocaleString()}</td>
                            </tr>
                            `;
                        }).join('')}
                    </tbody>
                </table>
                </div>
            </div>
            ` : '<div class="report-section"><p>No services assigned</p></div>'}

            ${report.ip_assignments && report.ip_assignments.length > 0 ? `
            <div class="report-section">
                <h3>IP Addresses</h3>
                <div class="table-wrapper">
                <table>
                    <thead>
                        <tr>
                            <th>Address</th>
                            <th>Type</th>
                            <th>Primary</th>
                            <th>Assigned</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${report.ip_assignments.map(ipa => {
                            const ip = ipMap[ipa.ip_id] || {};
                            return `
                            <tr>
                                <td><strong>${escapeHtml(ip.address || ipa.ip_id)}</strong></td>
                                <td>${escapeHtml(ip.type || '-')}</td>
                                <td>${ipa.is_primary ? '✓' : ''}</td>
                                <td>${new Date(ipa.created_at).toLocaleString()}</td>
                            </tr>
                            `;
                        }).join('')}
                    </tbody>
                </table>
                </div>
            </div>
            ` : '<div class="report-section"><p>No IP addresses assigned</p></div>'}

            ${report.journal_entries && report.journal_entries.length > 0 ? `
            <div class="report-section">
                <h3>Journal Entries</h3>
                <div class="table-wrapper">
                <table>
                    <thead>
                        <tr>
                            <th>Date</th>
                            <th>Category</th>
                            <th>Content</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${report.journal_entries.sort((a, b) => new Date(b.created_at) - new Date(a.created_at)).map(j => `
                            <tr>
                                <td>${new Date(j.created_at).toLocaleString()}</td>
                                <td>${escapeHtml(j.category)}</td>
                                <td class="wrap">${escapeHtml(j.content)}</td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
                </div>
            </div>
            ` : '<div class="report-section"><p>No journal entries</p></div>'}
        `;

        reportContent.innerHTML = html;
    } catch (error) {
        reportContent.innerHTML = `<div class="error">Error loading report: ${error.message}</div>`;
    }
}

// Modal Functions
function showCreateComputeModal() {
    const html = `
        <h2>Create Compute Resource</h2>
        <form id="create-compute-form">
            <div class="form-group">
                <label>Name *</label>
                <input type="text" name="name" required>
            </div>
            <div class="form-group">
                <label>Type *</label>
                <select name="type" required>
                    <option value="baremetal">Baremetal</option>
                    <option value="vps">VPS</option>
                    <option value="vm">VM</option>
                </select>
            </div>
            <div class="form-group">
                <label>Provider *</label>
                <input type="text" name="provider" required>
            </div>
            <div class="form-group">
                <label>Region *</label>
                <input type="text" name="region" required>
            </div>
            <div class="form-group">
                <label>Tags (format: key=value,key2=value2)</label>
                <input type="text" name="tags" placeholder="env=prod,zone=us-east">
            </div>
            <div class="form-actions">
                <button type="button" class="btn" onclick="closeModal()">Cancel</button>
                <button type="submit" class="btn btn-primary">Create</button>
            </div>
        </form>
    `;

    showModal(html);

    document.getElementById('create-compute-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const tags = parseTags(formData.get('tags'));

        const compute = {
            name: formData.get('name'),
            type: formData.get('type'),
            provider: formData.get('provider'),
            region: formData.get('region'),
            tags: tags,
            state: 'active'
        };

        try {
            await fetchAPI('/computes', 'POST', compute);
            closeModal();
            loadView('computes');
        } catch (error) {
            alert('Error creating compute: ' + error.message);
        }
    });
}

function showCreateComponentModal() {
    const html = `
        <h2>Create Component</h2>
        <form id="create-component-form">
            <div class="form-group">
                <label>Name *</label>
                <input type="text" name="name" required>
            </div>
            <div class="form-group">
                <label>Type *</label>
                <select name="type" required>
                    <option value="cpu">CPU</option>
                    <option value="ram">RAM</option>
                    <option value="storage">Storage</option>
                    <option value="gpu">GPU</option>
                    <option value="nic">NIC</option>
                    <option value="psu">PSU</option>
                    <option value="os">OS</option>
                    <option value="other">Other</option>
                </select>
            </div>
            <div class="form-group">
                <label>Manufacturer *</label>
                <input type="text" name="manufacturer" required>
            </div>
            <div class="form-group">
                <label>Model *</label>
                <input type="text" name="model" required>
            </div>
            <div class="form-group">
                <label>Specs (JSON)</label>
                <textarea name="specs" placeholder='{"cores":8,"ghz":3.5}'></textarea>
            </div>
            <div class="form-actions">
                <button type="button" class="btn" onclick="closeModal()">Cancel</button>
                <button type="submit" class="btn btn-primary">Create</button>
            </div>
        </form>
    `;

    showModal(html);

    document.getElementById('create-component-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);

        let specs = {};
        try {
            const specsStr = formData.get('specs');
            if (specsStr) {
                specs = JSON.parse(specsStr);
            }
        } catch (error) {
            alert('Invalid JSON in specs field');
            return;
        }

        const component = {
            name: formData.get('name'),
            type: formData.get('type'),
            manufacturer: formData.get('manufacturer'),
            model: formData.get('model'),
            specs: specs
        };

        try {
            await fetchAPI('/components', 'POST', component);
            closeModal();
            loadView('components');
        } catch (error) {
            alert('Error creating component: ' + error.message);
        }
    });
}

function showCreateServiceModal() {
    const html = `
        <h2>Create Service</h2>
        <form id="create-service-form">
            <div class="form-group">
                <label>Name *</label>
                <input type="text" name="name" required>
            </div>
            <div class="form-group">
                <label>Min Spec (JSON)</label>
                <textarea name="min_spec" placeholder='{"cores":2,"memory":4096}'></textarea>
            </div>
            <div class="form-group">
                <label>Max Spec (JSON)</label>
                <textarea name="max_spec" placeholder='{"cores":4,"memory":8192}'></textarea>
            </div>
            <div class="form-actions">
                <button type="button" class="btn" onclick="closeModal()">Cancel</button>
                <button type="submit" class="btn btn-primary">Create</button>
            </div>
        </form>
    `;

    showModal(html);

    document.getElementById('create-service-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);

        let min_spec = {};
        let max_spec = {};

        try {
            const minStr = formData.get('min_spec');
            const maxStr = formData.get('max_spec');
            if (minStr) min_spec = JSON.parse(minStr);
            if (maxStr) max_spec = JSON.parse(maxStr);
        } catch (error) {
            alert('Invalid JSON in spec fields');
            return;
        }

        const service = {
            name: formData.get('name'),
            min_spec: min_spec,
            max_spec: max_spec
        };

        try {
            await fetchAPI('/services', 'POST', service);
            closeModal();
            loadView('services');
        } catch (error) {
            alert('Error creating service: ' + error.message);
        }
    });
}

function showCreateIPModal() {
    const html = `
        <h2>Create IP Address</h2>
        <form id="create-ip-form">
            <div class="form-group">
                <label>Address *</label>
                <input type="text" name="address" required placeholder="192.168.1.100">
            </div>
            <div class="form-group">
                <label>Type *</label>
                <select name="type" required>
                    <option value="public">Public</option>
                    <option value="private">Private</option>
                </select>
            </div>
            <div class="form-group">
                <label>CIDR</label>
                <input type="text" name="cidr" placeholder="192.168.1.0/24">
            </div>
            <div class="form-group">
                <label>Gateway</label>
                <input type="text" name="gateway" placeholder="192.168.1.1">
            </div>
            <div class="form-actions">
                <button type="button" class="btn" onclick="closeModal()">Cancel</button>
                <button type="submit" class="btn btn-primary">Create</button>
            </div>
        </form>
    `;

    showModal(html);

    document.getElementById('create-ip-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);

        const ip = {
            address: formData.get('address'),
            type: formData.get('type'),
            cidr: formData.get('cidr') || '',
            gateway: formData.get('gateway') || ''
        };

        try {
            await fetchAPI('/ips', 'POST', ip);
            closeModal();
            loadView('ips');
        } catch (error) {
            alert('Error creating IP: ' + error.message);
        }
    });
}

function showCreateDNSModal() {
    const html = `
        <h2>Create DNS Record</h2>
        <form id="create-dns-form">
            <div class="form-group">
                <label>Name *</label>
                <input type="text" name="name" required placeholder="example.com">
            </div>
            <div class="form-group">
                <label>Type *</label>
                <select name="type" required>
                    <option value="A">A</option>
                    <option value="AAAA">AAAA</option>
                    <option value="CNAME">CNAME</option>
                    <option value="MX">MX</option>
                    <option value="TXT">TXT</option>
                    <option value="NS">NS</option>
                </select>
            </div>
            <div class="form-group">
                <label>Value *</label>
                <input type="text" name="value" required placeholder="1.2.3.4">
            </div>
            <div class="form-group">
                <label>Zone *</label>
                <input type="text" name="zone" required placeholder="example.com">
            </div>
            <div class="form-group">
                <label>TTL</label>
                <input type="number" name="ttl" value="3600">
            </div>
            <div class="form-actions">
                <button type="button" class="btn" onclick="closeModal()">Cancel</button>
                <button type="submit" class="btn btn-primary">Create</button>
            </div>
        </form>
    `;

    showModal(html);

    document.getElementById('create-dns-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);

        const dns = {
            name: formData.get('name'),
            type: formData.get('type'),
            value: formData.get('value'),
            zone: formData.get('zone'),
            ttl: parseInt(formData.get('ttl'))
        };

        try {
            await fetchAPI('/dns', 'POST', dns);
            closeModal();
            loadView('dns');
        } catch (error) {
            alert('Error creating DNS record: ' + error.message);
        }
    });
}

function showCreateFirewallModal() {
    const html = `
        <h2>Create Firewall Rule</h2>
        <form id="create-firewall-form">
            <div class="form-group">
                <label>Name *</label>
                <input type="text" name="name" required placeholder="allow-ssh">
            </div>
            <div class="form-group">
                <label>Action *</label>
                <select name="action" required>
                    <option value="ALLOW">ALLOW</option>
                    <option value="DENY">DENY</option>
                </select>
            </div>
            <div class="form-group">
                <label>Protocol *</label>
                <select name="protocol" required>
                    <option value="tcp">TCP</option>
                    <option value="udp">UDP</option>
                    <option value="icmp">ICMP</option>
                </select>
            </div>
            <div class="form-group">
                <label>Source *</label>
                <input type="text" name="source" required placeholder="any">
            </div>
            <div class="form-group">
                <label>Destination *</label>
                <input type="text" name="destination" required placeholder="any">
            </div>
            <div class="form-group">
                <label>Port Start</label>
                <input type="number" name="port_start" placeholder="22">
            </div>
            <div class="form-group">
                <label>Port End</label>
                <input type="number" name="port_end" placeholder="22">
            </div>
            <div class="form-actions">
                <button type="button" class="btn" onclick="closeModal()">Cancel</button>
                <button type="submit" class="btn btn-primary">Create</button>
            </div>
        </form>
    `;

    showModal(html);

    document.getElementById('create-firewall-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);

        const firewall = {
            name: formData.get('name'),
            action: formData.get('action'),
            protocol: formData.get('protocol'),
            source: formData.get('source'),
            destination: formData.get('destination'),
            port_start: parseInt(formData.get('port_start')) || 0,
            port_end: parseInt(formData.get('port_end')) || 0
        };

        try {
            await fetchAPI('/firewall', 'POST', firewall);
            closeModal();
            loadView('firewall');
        } catch (error) {
            alert('Error creating firewall rule: ' + error.message);
        }
    });
}

function showCreatePortModal() {
    const html = `
        <h2>Create Port Mapping</h2>
        <form id="create-port-form">
            <div class="form-group">
                <label>Assignment ID *</label>
                <input type="text" name="assignment_id" required>
            </div>
            <div class="form-group">
                <label>IP ID *</label>
                <input type="text" name="ip_id" required>
            </div>
            <div class="form-group">
                <label>External Port *</label>
                <input type="number" name="port" required placeholder="8080">
            </div>
            <div class="form-group">
                <label>Protocol *</label>
                <select name="protocol" required>
                    <option value="tcp">TCP</option>
                    <option value="udp">UDP</option>
                </select>
            </div>
            <div class="form-group">
                <label>Service Port *</label>
                <input type="number" name="service_port" required placeholder="80">
            </div>
            <div class="form-actions">
                <button type="button" class="btn" onclick="closeModal()">Cancel</button>
                <button type="submit" class="btn btn-primary">Create</button>
            </div>
        </form>
    `;

    showModal(html);

    document.getElementById('create-port-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);

        const port = {
            assignment_id: formData.get('assignment_id'),
            ip_id: formData.get('ip_id'),
            port: parseInt(formData.get('port')),
            protocol: formData.get('protocol'),
            service_port: parseInt(formData.get('service_port'))
        };

        try {
            await fetchAPI('/ports', 'POST', port);
            closeModal();
            loadView('ports');
        } catch (error) {
            alert('Error creating port mapping: ' + error.message);
        }
    });
}

function showCreateJournalModal() {
    const html = `
        <h2>Create Journal Entry</h2>
        <form id="create-journal-form">
            <div class="form-group">
                <label>Compute ID *</label>
                <input type="text" name="compute_id" required>
            </div>
            <div class="form-group">
                <label>Category *</label>
                <select name="category" required>
                    <option value="maintenance">Maintenance</option>
                    <option value="deployment">Deployment</option>
                    <option value="incident">Incident</option>
                    <option value="change">Change</option>
                    <option value="other">Other</option>
                </select>
            </div>
            <div class="form-group">
                <label>Content *</label>
                <textarea name="content" required placeholder="Journal entry details"></textarea>
            </div>
            <div class="form-actions">
                <button type="button" class="btn" onclick="closeModal()">Cancel</button>
                <button type="submit" class="btn btn-primary">Create</button>
            </div>
        </form>
    `;

    showModal(html);

    document.getElementById('create-journal-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);

        const journal = {
            compute_id: formData.get('compute_id'),
            category: formData.get('category'),
            content: formData.get('content')
        };

        try {
            await fetchAPI('/journal', 'POST', journal);
            closeModal();
            loadView('journal');
        } catch (error) {
            alert('Error creating journal entry: ' + error.message);
        }
    });
}

function showCreateAPIKeyModal() {
    const html = `
        <h2>Create API Key</h2>
        <form id="create-apikey-form">
            <div class="form-group">
                <label>Name *</label>
                <input type="text" name="name" required placeholder="developer-key">
            </div>
            <div class="form-group">
                <label>Scope *</label>
                <select name="scope" required>
                    <option value="admin">Admin</option>
                    <option value="readwrite">Read/Write</option>
                    <option value="readonly">Read Only</option>
                </select>
            </div>
            <div class="form-group">
                <label>Description</label>
                <textarea name="description" placeholder="Optional description"></textarea>
            </div>
            <div class="form-actions">
                <button type="button" class="btn" onclick="closeModal()">Cancel</button>
                <button type="submit" class="btn btn-primary">Create</button>
            </div>
        </form>
    `;

    showModal(html);

    document.getElementById('create-apikey-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);

        const apikey = {
            name: formData.get('name'),
            scope: formData.get('scope'),
            description: formData.get('description') || ''
        };

        try {
            const result = await fetchAPI('/admin/apikeys', 'POST', apikey);
            alert('API Key created:\n\n' + result.key + '\n\nSave this key - it will not be shown again!');
            closeModal();
            loadView('apikeys');
        } catch (error) {
            alert('Error creating API key: ' + error.message);
        }
    });
}

function showModal(html) {
    const modal = document.getElementById('modal');
    const modalBody = document.getElementById('modal-body');
    modalBody.innerHTML = html;
    modal.classList.remove('hidden');
    modal.classList.add('show');
}

function closeModal() {
    const modal = document.getElementById('modal');
    modal.classList.remove('show');
    modal.classList.add('hidden');
}

// Close modal on click outside or X
document.addEventListener('click', (e) => {
    const modal = document.getElementById('modal');
    if (e.target === modal || e.target.classList.contains('close')) {
        closeModal();
    }
});

// Delete Functions
async function deleteCompute(id) {
    if (!confirm('Delete this compute resource?')) return;
    try {
        await fetchAPI(`/computes/${id}`, 'DELETE');
        loadView('computes');
    } catch (error) {
        alert('Error deleting compute: ' + error.message);
    }
}

async function deleteComponent(id) {
    if (!confirm('Delete this component?')) return;
    try {
        await fetchAPI(`/components/${id}`, 'DELETE');
        loadView('components');
    } catch (error) {
        alert('Error deleting component: ' + error.message);
    }
}

async function deleteService(id) {
    if (!confirm('Delete this service?')) return;
    try {
        await fetchAPI(`/services/${id}`, 'DELETE');
        loadView('services');
    } catch (error) {
        alert('Error deleting service: ' + error.message);
    }
}

async function deleteIP(id) {
    if (!confirm('Delete this IP address?')) return;
    try {
        await fetchAPI(`/ips/${id}`, 'DELETE');
        loadView('ips');
    } catch (error) {
        alert('Error deleting IP: ' + error.message);
    }
}

async function deleteDNS(id) {
    if (!confirm('Delete this DNS record?')) return;
    try {
        await fetchAPI(`/dns/${id}`, 'DELETE');
        loadView('dns');
    } catch (error) {
        alert('Error deleting DNS record: ' + error.message);
    }
}

async function deleteFirewall(id) {
    if (!confirm('Delete this firewall rule?')) return;
    try {
        await fetchAPI(`/firewall/${id}`, 'DELETE');
        loadView('firewall');
    } catch (error) {
        alert('Error deleting firewall rule: ' + error.message);
    }
}

async function deletePort(id) {
    if (!confirm('Delete this port mapping?')) return;
    try {
        await fetchAPI(`/ports/${id}`, 'DELETE');
        loadView('ports');
    } catch (error) {
        alert('Error deleting port mapping: ' + error.message);
    }
}

async function deleteAPIKey(id) {
    if (!confirm('Delete this API key?')) return;
    try {
        await fetchAPI(`/admin/apikeys/${id}`, 'DELETE');
        loadView('apikeys');
    } catch (error) {
        alert('Error deleting API key: ' + error.message);
    }
}

// API Helper
async function fetchAPI(endpoint, method = 'GET', body = null) {
    const options = {
        method,
        headers: {
            'Content-Type': 'application/json'
        }
    };

    if (body) {
        options.body = JSON.stringify(body);
    }

    const response = await fetch(API_BASE + endpoint, options);

    if (!response.ok) {
        const text = await response.text();
        throw new Error(text || response.statusText);
    }

    return response.json();
}

// Utility Functions
function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function renderTags(tags) {
    if (!tags || Object.keys(tags).length === 0) return '-';
    return Object.entries(tags)
        .map(([k, v]) => `<span class="tag">${escapeHtml(k)}=${escapeHtml(v)}</span>`)
        .join('');
}

function parseTags(str) {
    if (!str) return {};
    const tags = {};
    str.split(',').forEach(pair => {
        const [key, value] = pair.split('=').map(s => s.trim());
        if (key && value) {
            tags[key] = value;
        }
    });
    return tags;
}
