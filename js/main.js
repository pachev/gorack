var app = new Vue({
    //TODO: Add a means of fetching using user provided available weights
    el: '#app',
    data: {
        barSelect: 'olympicBar',
        desiredWeight: 185,
        barWeights: {
            'ezBar': 25,
            'shortBar': 35,
            'olympicBar': 45
        },
        defaults: {
            fortyFives: 10,
            thirtyFives: 10,
            twentyFives: 10,
            tens: 10,
            fives: 10,
            twoDotFives: 10,
        },
        results: {},
        url: 'https://gorack.pachevjoseph.com/v1/api/rack',
        hasError: false
    },
    methods: {
        selectBar(bar) {
            this.barSelect = bar;
        },
        calculateWeight() {
            this.hasError = false;
            axios.post(this.url, {
                    ...this.defaults,
                    barWeight: this.barWeights[this.barSelect],
                    desiredWeight: this.desiredWeight
                })
                .then((response) => {
                    this.results = response.data;
                })
                .catch((error) => {
                    this.hasError = true;
                    console.log(error);
                });
        }
    }
})