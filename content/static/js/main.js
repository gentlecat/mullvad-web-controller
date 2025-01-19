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
        button.textContent = `${flagEmoji} ${location.city.toUpperCase()}`;

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
            {country: "al", city: "tia"}, // Tirana, Albania
            {country: "ca", city: "yvn"}, // Vancouver, Canada
            {country: "ch", city: "zrh"}, // Zurich, Switzerland
            {country: "de", city: "fra"}, // Frankfurt, Germany
            {country: "jp", city: "tyo"}, // Tokyo, Japan
            {country: "nl", city: "ams"}, // Amsterdam, Netherlands
            {country: "se", city: "sth"}, // Stockholm, Sweden
            {country: "sg", city: "sin"}, // Singapore
            {country: "tr", city: "ist"}, // Istanbul, Turkey
            {country: "us", city: "sea"}, // Seattle, USA
            {country: "us", city: "sjc"}, // San Jose, USA
        ]
    );

};
