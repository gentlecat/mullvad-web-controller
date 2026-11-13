const STATUS_REFRESH_INTERVAL = 15000;
const SWITCH_POLL_INTERVAL = 1000;
const SWITCH_TIMEOUT = 30000;
const MAX_RECENT = 4;
const RECENT_STORAGE_KEY = 'recentLocations';

const STATE_LABELS = {
    connected: 'Connected',
    connecting: 'Connecting',
    disconnected: 'Disconnected',
    disconnecting: 'Disconnecting',
    error: 'Blocked, the tunnel has an error',
};

const elements = {};

let relays = [];
let tunnelStatus = null;
let pendingTarget = null;
let statusTimer = null;

// Helpers

function flagEmoji(countryCode) {
    if (!countryCode || countryCode.length !== 2) {
        return '';
    }
    return String.fromCodePoint(...countryCode.toUpperCase().split('').map(char => 0x1F1E6 + char.charCodeAt(0) - 65));
}

// Lowercase and strip accents, so that "sao" finds "São Paulo"
function normalize(text) {
    return text.normalize('NFD').replace(/\p{Diacritic}/gu, '').toLowerCase();
}

function element(tag, className, text) {
    const node = document.createElement(tag);
    if (className) {
        node.className = className;
    }
    if (text !== undefined) {
        node.textContent = text;
    }
    return node;
}

function countryName(countryCode) {
    return relays.find(relay => relay.country === countryCode)?.countryName ?? countryCode.toUpperCase();
}

// Returns a human-readable name for a location, such as "Gothenburg, Sweden" or "Any city in Sweden"
function locationName({country, city}) {
    if (!city) {
        return `Any city in ${countryName(country)}`;
    }
    const relay = relays.find(relay => relay.country === country && relay.city === city);
    return relay ? `${relay.name}, ${relay.countryName}` : countryName(country);
}

function sameLocation(a, b) {
    return a && b && a.country === b.country && (a.city ?? '') === (b.city ?? '');
}

// Whether the tunnel is connected to the target. A country-only target matches any city in it.
function statusMatches(status, target) {
    return status?.state === 'connected' && status.country === target.country && (!target.city || status.city === target.city);
}

async function fetchJSON(url, options) {
    const response = await fetch(url, options);
    if (!response.ok) {
        const text = (await response.text()).trim();
        throw new Error(text || `Request failed with status ${response.status}`);
    }
    return response.json();
}

// Recent locations, kept in this browser only

function loadRecent() {
    try {
        const recent = JSON.parse(localStorage.getItem(RECENT_STORAGE_KEY));
        return Array.isArray(recent) ? recent.filter(item => typeof item?.country === 'string') : [];
    } catch {
        return [];
    }
}

function saveRecent(target) {
    const recent = [target, ...loadRecent().filter(item => !sameLocation(item, target))].slice(0, MAX_RECENT);
    try {
        localStorage.setItem(RECENT_STORAGE_KEY, JSON.stringify(recent));
    } catch {
        // Storage can be unavailable, for example in private browsing. Recent locations are a convenience.
    }
    renderRecent();
}

// Rendering

function showNotice(message, retry) {
    elements.noticeText.textContent = message;
    elements.noticeAction.hidden = !retry;
    elements.noticeAction.onclick = retry ?? null;
    elements.notice.hidden = false;
}

function hideNotice() {
    elements.notice.hidden = true;
}

function renderStatus() {
    let state;
    let flag;
    let location;
    let details;

    if (pendingTarget) {
        state = 'switching';
        flag = flagEmoji(pendingTarget.country);
        location = locationName(pendingTarget);
        details = 'Waiting for the tunnel to reconnect';
    } else if (!tunnelStatus) {
        state = 'unavailable';
        flag = '';
        location = 'Status unavailable';
        details = 'Could not reach the Mullvad daemon';
    } else {
        state = tunnelStatus.state;
        flag = flagEmoji(tunnelStatus.country);
        // Prefer the names the daemon reports, and fall back to the relay list
        location = [tunnelStatus.cityName, tunnelStatus.countryName].filter(Boolean).join(', ')
            || (tunnelStatus.country ? locationName(tunnelStatus) : 'Unknown location');
        details = [tunnelStatus.hostname, tunnelStatus.ipv4].filter(Boolean).join(' · ');
    }

    elements.status.dataset.state = state;
    elements.statusFlag.textContent = flag;
    elements.statusState.textContent = state === 'switching' ? 'Switching' : STATE_LABELS[state] ?? 'Unknown';
    elements.statusLocation.textContent = location;
    elements.statusDetails.textContent = details;

    renderMarkers();
}

// Marks the connected and pending locations in the list and the recent chips
function renderMarkers() {
    const current = tunnelStatus?.state === 'connected' ? tunnelStatus : null;

    for (const button of document.querySelectorAll('[data-country]')) {
        const target = {country: button.dataset.country, city: button.dataset.city};
        const isPending = sameLocation(pendingTarget, target);
        const isCurrent = !pendingTarget && current && target.city && sameLocation(current, target);

        button.setAttribute('aria-current', isCurrent ? 'true' : 'false');
        button.disabled = Boolean(pendingTarget);

        const marker = button.querySelector('.location-marker');
        if (marker) {
            marker.textContent = isPending ? 'Connecting' : isCurrent ? 'Connected' : '';
            marker.classList.toggle('pending', isPending);
        }
    }

    elements.locationList.setAttribute('aria-busy', pendingTarget ? 'true' : 'false');
}

function locationButton({country, city, className, flag, name, hint}) {
    const button = element('button', `location ${className}`);
    button.type = 'button';
    button.dataset.country = country;
    button.dataset.city = city ?? '';
    button.append(
        element('span', 'location-flag', flag ?? ''),
        element('span', 'location-name', name),
    );
    if (hint) {
        button.querySelector('.location-name').append(' ', element('span', 'location-hint', hint));
    }
    button.append(element('span', 'location-marker'));
    return button;
}

function renderLocations() {
    // Relays arrive sorted by country, so grouping keeps that order
    const countries = new Map();
    for (const relay of relays) {
        countries.set(relay.country, [...(countries.get(relay.country) ?? []), relay]);
    }
    const groups = [];

    for (const [country, cities] of countries) {
        const group = element('div', 'country');
        const name = cities[0].countryName;
        group.dataset.search = normalize(`${name} ${country}`);

        if (cities.length === 1) {
            // A country with one city gets a single row
            const [relay] = cities;
            const button = locationButton({
                country, city: relay.city, className: 'country-name', flag: flagEmoji(country),
                name: relay.countryName, hint: relay.name,
            });
            button.dataset.search = normalize(`${relay.name} ${relay.city}`);
            group.append(button);
        } else {
            group.append(locationButton({
                country, className: 'country-name', flag: flagEmoji(country), name, hint: 'Any city',
            }));
            for (const relay of cities) {
                const button = locationButton({country, city: relay.city, className: 'city', name: relay.name});
                button.dataset.search = normalize(`${relay.name} ${relay.city}`);
                group.append(button);
            }
        }
        groups.push(group);
    }

    elements.locationList.replaceChildren(...groups, element('p', 'placeholder', 'No locations match your search'));
    elements.search.disabled = false;
    filterLocations();
    renderMarkers();
}

function renderRecent() {
    const recent = relays.length ? loadRecent() : [];
    const chips = recent.map(target => {
        const chip = element('button', 'chip');
        chip.type = 'button';
        chip.dataset.country = target.country;
        chip.dataset.city = target.city ?? '';
        const relay = relays.find(relay => relay.country === target.country && relay.city === target.city);
        chip.append(
            element('span', '', flagEmoji(target.country)),
            element('span', '', relay?.name ?? countryName(target.country)),
        );
        chip.title = locationName(target);
        return chip;
    });

    elements.recentList.replaceChildren(...chips);
    elements.recentSection.hidden = chips.length === 0;
    renderMarkers();
}

function filterLocations() {
    const query = normalize(elements.search.value.trim());
    let anyVisible = false;

    for (const group of elements.locationList.querySelectorAll('.country')) {
        // If the country matches, show all of its cities
        const countryMatches = group.dataset.search.includes(query);
        let groupVisible = false;
        for (const button of group.querySelectorAll('.location')) {
            const visible = !query || countryMatches || (button.dataset.search ?? '').includes(query);
            button.hidden = !visible;
            groupVisible ||= visible;
        }
        group.hidden = !groupVisible;
        anyVisible ||= groupVisible;
    }

    elements.locationList.querySelector('.placeholder').hidden = anyVisible;
}

// Actions

async function refreshStatus() {
    try {
        tunnelStatus = await fetchJSON('/api/status');
    } catch (error) {
        console.error('Error fetching the tunnel status:', error);
        tunnelStatus = null;
    }
    renderStatus();
    return tunnelStatus;
}

function scheduleStatusRefresh() {
    clearTimeout(statusTimer);
    if (!document.hidden) {
        statusTimer = setTimeout(async () => {
            await refreshStatus();
            scheduleStatusRefresh();
        }, STATUS_REFRESH_INTERVAL);
    }
}

async function switchTo(target) {
    if (pendingTarget) {
        return;
    }

    hideNotice();
    clearTimeout(statusTimer);
    pendingTarget = target;
    renderStatus();

    try {
        await fetchJSON('/api/relay/location', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(target.city ? target : {country: target.country}),
        });
        saveRecent(target);

        // Poll until the tunnel reports the new location
        const deadline = Date.now() + SWITCH_TIMEOUT;
        let status = null;
        while (Date.now() < deadline) {
            await new Promise(resolve => setTimeout(resolve, SWITCH_POLL_INTERVAL));
            status = await fetchJSON('/api/status').catch(() => null);
            if (statusMatches(status, target)) {
                break;
            }
        }

        tunnelStatus = status;
        if (!statusMatches(status, target)) {
            showNotice(`Still not connected to ${locationName(target)}. The tunnel may need more time.`);
        }
    } catch (error) {
        console.error('Error changing location:', error);
        showNotice(`Could not switch to ${locationName(target)}: ${error.message}`, () => switchTo(target));
    } finally {
        pendingTarget = null;
        renderStatus();
        scheduleStatusRefresh();
    }
}

async function loadLocations() {
    hideNotice();
    try {
        relays = await fetchJSON('/api/relays');
    } catch (error) {
        console.error('Error fetching locations:', error);
        elements.locationList.replaceChildren(element('p', 'placeholder', 'Locations are unavailable'));
        showNotice(`Could not load locations: ${error.message}`, loadLocations);
        return;
    }
    renderLocations();
    renderRecent();
    renderStatus();
}

function handleLocationClick(event) {
    const button = event.target.closest('[data-country]');
    if (button && !button.disabled) {
        switchTo({country: button.dataset.country, city: button.dataset.city || undefined});
    }
}

function init() {
    for (const id of ['status', 'notice', 'search']) {
        elements[id] = document.getElementById(id);
    }
    elements.statusFlag = document.getElementById('status-flag');
    elements.statusState = document.getElementById('status-state');
    elements.statusLocation = document.getElementById('status-location');
    elements.statusDetails = document.getElementById('status-details');
    elements.noticeText = document.getElementById('notice-text');
    elements.noticeAction = document.getElementById('notice-action');
    elements.recentSection = document.getElementById('recent-section');
    elements.recentList = document.getElementById('recent-list');
    elements.locationList = document.getElementById('location-list');

    elements.locationList.addEventListener('click', handleLocationClick);
    elements.recentList.addEventListener('click', handleLocationClick);

    elements.search.addEventListener('input', filterLocations);
    elements.search.addEventListener('keydown', event => {
        if (event.key === 'Enter') {
            // Pick the first match, so that typing a city and pressing Enter switches to it
            elements.locationList.querySelector('.location:not([hidden])')?.click();
        } else if (event.key === 'Escape') {
            elements.search.value = '';
            filterLocations();
        }
    });

    document.addEventListener('keydown', event => {
        if (event.key === '/' && document.activeElement !== elements.search && !elements.search.disabled) {
            event.preventDefault();
            elements.search.focus();
        }
    });

    document.addEventListener('visibilitychange', async () => {
        if (!document.hidden && !pendingTarget) {
            await refreshStatus();
        }
        scheduleStatusRefresh();
    });

    loadLocations();
    refreshStatus().then(scheduleStatusRefresh);
}

document.addEventListener('DOMContentLoaded', init);
