let choicesInstance = null;

function initLocations(selectElement, locations) {
    selectElement.innerHTML = '<option value="">Choose a location...</option>';

    // Transform locations into choices format
    const choices = locations.map(location => {
        const flagEmoji = String.fromCodePoint(
            ...location.country
                .toUpperCase()
                .split('')
                .map(char => 0x1F1E6 + char.charCodeAt(0) - 65)
        );
        return {
            value: JSON.stringify({country: location.country, city: location.city}),
            label: `${flagEmoji} ${location.name}`,
            customProperties: {
                country: location.country,
                city: location.city
            }
        };
    });

    choicesInstance = new Choices(selectElement, {
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
            setResponseMessage('<i>Connecting to selected location...</i>');

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
                        setResponseMessage('<i>Retrieving the current location...</i>');
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
                setResponseMessage("Failed to fetch current IP info");
                return;
            }
            return response.json();
        })
        .then(ipInfo => {
            setResponseMessage(`
                Country: <b>${ipInfo.country}</b><br />
                City: <b>${ipInfo.city}</b><br />
                IP: ${ipInfo.ip}<br />
                Hosted by: ${ipInfo.org}
            `)
        })
        .catch(error => {
            setResponseMessage(`Error: ${error.message}`);
            console.error('Error fetching IP info:', error);
        });
}

function setResponseMessage(html) {
    document.getElementById('response').innerHTML = html;
}

function init() {
    // Hide the location selector and show loading message
    const locationContainer = document.getElementById('location-selector').parentElement;
    locationContainer.style.display = 'none';
    setResponseMessage('<i>Loading locations...</i>');

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
