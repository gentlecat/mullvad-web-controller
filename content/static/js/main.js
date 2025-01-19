function initLocations(locationsContainer, locations) {

    locations.forEach(location => {
        const button = document.createElement('button');
        button.classList.add('btn', 'btn-outline-primary');

        const flagEmoji = String.fromCodePoint(
            ...location.country
                .toUpperCase()
                .split('')
                .map(char => 0x1F1E6 + char.charCodeAt(0) - 65)
        );
        button.textContent = `${flagEmoji} ${location.name}`;

        button.setAttribute('hx-post', '/api/relay/location');
        button.setAttribute('hx-ext', 'json-enc');
        button.setAttribute('hx-vals', JSON.stringify({country: location.country, city: location.city}));
        button.setAttribute('hx-target', '#response');
        button.setAttribute('hx-swap', 'none');

        locationsContainer.appendChild(button);

        button.addEventListener('htmx:afterRequest', function (event) {
            if (event.detail.successful) {
                console.debug(event.detail.xhr);
                setCurrentLocation();
            } else {
                setResponseMessage(event.detail.xhr.response);
            }
        });
    });

    htmx.process(locationsContainer);
}

function setCurrentLocation() {
    fetch('/api/ip')
        .then(response => {
            if (!response.ok) {
                setResponseMessage("Failed to fetch current IP info");
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
            setResponseMessage(error);
        });
}

function setResponseMessage(html) {
    document.getElementById('response').innerHTML = html;
}

window.onload = function () {

    initLocations(
        document.getElementById('location-selector'),
        [
            {country: "al", city: "tia", name: "Tirana"},
            {country: "ca", city: "van", name: "Vancouver"},
            {country: "ch", city: "zrh", name: "Zurich"},
            {country: "de", city: "fra", name: "Frankfurt"},
            {country: "jp", city: "tyo", name: "Tokyo"},
            {country: "nl", city: "ams", name: "Amsterdam"},
            {country: "se", city: "sth", name: "Stockholm"},
            {country: "sg", city: "sin", name: "Singapore"},
            {country: "tr", city: "ist", name: "Istanbul"},
            {country: "us", city: "sea", name: "Seattle"},
            {country: "us", city: "sjc", name: "San Jose"},
        ]
    );

    setCurrentLocation();

};
