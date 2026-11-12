let choicesInstance = null;

// Follow the system color scheme, including changes while the page is open
const darkSchemeQuery = window.matchMedia('(prefers-color-scheme: dark)');

function applyColorScheme() {
    document.documentElement.setAttribute('data-bs-theme', darkSchemeQuery.matches ? 'dark' : 'light');
}

applyColorScheme();
darkSchemeQuery.addEventListener('change', applyColorScheme);

function initLocations(selectElement, locations) {
    selectElement.replaceChildren(new Option('Choose a location...', ''));

    // Transform locations into choices format
    const choices = locations.map(location => {
        const flagEmoji = String.fromCodePoint(
            ...location.country
                .toUpperCase()
                .split('')
                .map(char => 0x1F1E6 + char.charCodeAt(0) - 65)
        );
        const name = location.countryName ? `${location.name}, ${location.countryName}` : location.name;
        return {
            value: JSON.stringify({country: location.country, city: location.city}),
            label: `${flagEmoji} ${name}`,
            customProperties: {
                country: location.country,
                city: location.city
            }
        };
    });

    choicesInstance = new Choices(selectElement, {
        allowHTML: false,
        searchEnabled: true,
        searchPlaceholderValue: 'Search locations...',
        itemSelectText: 'Press to select',
        noResultsText: 'No locations found',
        noChoicesText: 'No locations available',
        choices: choices,
        shouldSort: false
    });

    selectElement.addEventListener('change', function(event) {
        const selectedValue = event.target.value;

        if (selectedValue) {
            const locationData = JSON.parse(selectedValue);
            setResponseMessage(italic('Connecting to selected location...'));

            fetch('/api/relay/location', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(locationData)
            })
            .then(response => {
                if (response.ok) {
                    setTimeout(() => {
                        setResponseMessage(italic('Retrieving the current location...'));
                        setCurrentLocation();
                    }, 2000);
                } else {
                    return response.text().then(text => {
                        throw new Error(text || 'Failed to change location');
                    });
                }
            })
            .catch(error => {
                setResponseMessage(`Error: ${error.message}`);
                console.error('Error changing location:', error);
            });
        }
    });
}

function setCurrentLocation() {
    fetch('/api/ip')
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to fetch current IP info');
            }
            return response.json();
        })
        .then(ipInfo => {
            setResponseMessage(
                'Country: ', bold(ipInfo.country), document.createElement('br'),
                'City: ', bold(ipInfo.city), document.createElement('br'),
                `IP: ${ipInfo.ip}`, document.createElement('br'),
                `Hosted by: ${ipInfo.org}`
            );
        })
        .catch(error => {
            setResponseMessage(`Error: ${error.message}`);
            console.error('Error fetching IP info:', error);
        });
}

// Accepts strings (inserted as text, never parsed as HTML) and DOM nodes
function setResponseMessage(...content) {
    document.getElementById('response').replaceChildren(...content);
}

function italic(text) {
    const element = document.createElement('i');
    element.textContent = text;
    return element;
}

function bold(text) {
    const element = document.createElement('b');
    element.textContent = text;
    return element;
}

function init() {
    // Hide the location selector and show loading message
    const locationContainer = document.getElementById('location-selector').parentElement;
    locationContainer.style.display = 'none';
    setResponseMessage(italic('Loading locations...'));

    fetch('/api/relays')
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to fetch locations');
            }
            return response.json();
        })
        .then(locations => {
            console.debug(locations);
            // Show the location selector
            locationContainer.style.display = 'block';
            initLocations(
                document.getElementById('location-selector'),
                locations
            );
            setCurrentLocation();
        })
        .catch(error => {
            console.error('Error fetching locations:', error);
            setResponseMessage(`Error: ${error.message}`);
            locationContainer.style.display = 'block';
        });
}

window.onload = function () {
    init();
};
