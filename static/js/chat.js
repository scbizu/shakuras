Vue.use(VueMaterial)

var conn;

new Vue({
  el:"#sidebar",
  data:{

  },

})

var chats = new Vue({
  el:"#chatscreen",
  data:{
      contents:[

      ],
  },
})


new Vue({
  el:"#inputbar",
  data:{
    msg:"",
  },
  methods:{
    send: function(){
        if(!conn){
          console.error("conn was not inited.");
        }
        if(!this.msg) {
          console.error("plz input sth..")
        }
        conn.send(this.msg);
        this.msg ="";
    }
  },
})

if (window["WebSocket"]) {
    conn = new WebSocket("ws://localhost:8090/ws")
    //conn close callback
    conn.onclose = function(e){
      chats.contents.push("user quited")
    }

    conn.onmessage = function(e){
      chats.contents.push(e.data)
    }
}else{
  //do not support WebSocket
  console.log("do not support WebSocket")
}
