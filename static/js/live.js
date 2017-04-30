Vue.use(VueMaterial)



var conn;

var player = document.getElementById("player");


var siderbar = new Vue({
  el:"#sidebar",
  data:{
    tags:[

    ],
  },
  beforeCreate :function(){
    fetch('/videotags')
    .then(function(response){
      response.json().then(function(data){
        siderbar.tags = data
      })
    })
    .catch(function(e){
      console.log(e)
    })
  },
  methods:{
    toggleSidenav:function(){
      this.$refs.leftnav.toggle()
    },
    tagshref : function(tagname){
      //
      return "/watch/"+tagname
    },
    // getSeries:function(tag){
    //   fetch('/getSeries?tagname='+tag)
    //   .then(function(response){
    //     response.json().then(function(data){
    //       console.log(data)
    //       series.parts = data
    //       series.hasContent = true
    //     })
    //   })
    //   // console.log("clicked!")
    //   // sideBarRouter.push({path:"/watch/"+tagname})
    // },
  }
})



// new Vue({
//   el:"#about",
//   data:{},
//   methods:{},
// })

var chats = new Vue({
  el:"#screen",
  data:{
      contents:[

      ],
      ldProgress: 0,
      msg:"",
      qa: "raw",
      isplay:false,
  },

  methods:{


    fullScreen:function(){


        if (player.requestFullscreen) {
        player.requestFullscreen();
      } else if (player.mozRequestFullScreen) {
        player.mozRequestFullScreen();
      } else if (player.webkitRequestFullscreen) {
        player.webkitRequestFullscreen();
      }
    },

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
});


// //fetch rsc


if (flvjs.isSupported()) {
    var vsrc = document.querySelector('video');
    var flvPlayer = flvjs.createPlayer({
        type: 'flv',
        url: 'http://localhost:8091/test'
    });
    flvPlayer.attachMediaElement(vsrc);
    flvPlayer.load();
    flvPlayer.play();
}

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
