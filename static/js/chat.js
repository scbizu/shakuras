Vue.use(VueMaterial)

var conn;

new Vue({
  el:"#sidebar",
  data:{

  },

})

var chats = new Vue({
  el:"#screen",
  data:{
      contents:[

      ],
  },

  methods:{

    playORpause:function(){
      var player = document.getElementById("player");

      if(player.paused){
        player.play();
      }else{
        player.pause();
      }
    },

    fullScreen:function(){
      var player = document.getElementById("player");

        if (player.requestFullscreen) {
        player.requestFullscreen();
      } else if (player.mozRequestFullScreen) {
        player.mozRequestFullScreen();
      } else if (player.webkitRequestFullscreen) {
        player.webkitRequestFullscreen();
      }
    },
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
      chats.contents.push("聊天室网线被踹断啦(PД`q。=)>·。'゜")
    }

    conn.onmessage = function(e){
      chats.contents.push(e.data)
    }
}else{
  //do not support WebSocket
  console.log("do not support WebSocket")
}
