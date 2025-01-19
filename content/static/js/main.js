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
                const responseData = JSON.parse(event.detail.xhr.response);
                const ipInfo = responseData['ip_info'];
                setResponseMessage(`IP: <b>${ipInfo.ip}</b><br />Country: <b>${ipInfo.country}</b><br />City: <b>${ipInfo.city}</b><br />Hosted by: <b>${ipInfo.org}</b>`);
            } else {
                setResponseMessage(event.detail.xhr.response);
            }
        });
    });

    htmx.process(locationsContainer);
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

};
