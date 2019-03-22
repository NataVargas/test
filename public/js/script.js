new Vue({
  el: 'body',

  data: {
    pairkeys: [],
    text_encrypt: [],
    text_plain: [],
    newpairkey: {}
  },

  // This is run whenever the page is loaded to make sure we have a current pairkey list
  created: function() {
    // Use the vue-resource $http client to fetch data from the /pairkeys route
    this.$http.get('http://localhost:3333/get').then(function(response) {
      for (i = 0; i < response.data.length ; i++) {
        console.log(response.data[i])
        this.pairkeys.push(response.data[i])
      }
    })
  },

  methods: {
    createpairkey: function() {
      if (!$.trim(this.newpairkey.name)) {
        this.newpairkey = {}
        return
      }

      // Post the new pairkey to the /pairkeys route using the $http client
      this.$http.post('http://localhost:3333/post/'+this.newpairkey.name).success(function(response) {
        this.pairkeys = []
        this.$http.get('http://localhost:3333/get').then(function(response) {
          for (i = 0; i < response.data.length ; i++) {
            console.log(response.data[i])
            this.pairkeys.push(response.data[i])
          }
        })
      }).error(function(error) {
        console.log(error)
      });
    },
    searchpairkey: function() {
      // Use the vue-resource $http client to fetch data from the /pairkeys route
      this.pairkeys = []
      this.$http.get('http://localhost:3333/get/'+ this.pairkey.name).then(function(response) {
        for (i = 0; i < response.data.length ; i++) {
          console.log(response.data[i])
          this.pairkeys.push(response.data[i])
        }
      })
    },

    encrypt: function() {
      this.text_encrypt = []
      // Use the vue-resource $http client to fetch data from the /pairkeys route
      var str = this.original.text;
      var res = str.split(" ");
      var text = "";
      console.log(res)

      for (i = 0; i < res.length ; i++) {
        if (i < res.length-1){
          text = text + res[i] + "_" ;
        } else {
          text = text + res[i];
        }
      }

      this.$http.get('http://localhost:3333/get/'+ this.pairkey.id + "/" + text).then(function(response) {
        this.text_encrypt.push(response.data);
        console.log("Texto cifrado: " + response.data);
      })
    },
    descrypt: function() {
      // Use the vue-resource $http client to fetch data from the /pairkeys route
      this.text_plain = [];
      var str = this.original2.text;
      var res = str.split(" ");
      var text = "";
      console.log(res)

      for (i = 0; i < res.length ; i++) {
        if (i < res.length-1){
          text = text + res[i] + "_" ;
        } else {
          text = text + res[i];
        }
      }
      this.$http.get('http://localhost:3333/getplain/'+ this.pairkey2.id + "/" + text).then(function(response) {
        this.text_plain.push(response.data);
        console.log("Texto original: " + response.data);
      })
    }
  }
})
