Vue.config.delimiters = ['(%', '%)']

var app = new Vue({
    el: '#maincontent',
    data: {
        pluginname: 'Plugin Starter',
        version: '0.1.0',
        description: 'The beginnings of yet another awesome plugin.'
    },
    computed: {
        shortname: function () {
            return this.pluginname.toLowerCase().replace(/\s/g, '-')
        }
    }
})
