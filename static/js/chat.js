Vue.use(VueMaterial)

var conn;

var player = document.getElementById("player");


new Vue({
  el:"#sidebar",
  data:{

  },
  methods:{
    toggleSidenav:function(){
      this.$refs.leftnav.toggle()
    }
  }
})

new Vue({
  el:"#series",
  data:{
    parts:[
      {vid:"0x00",vname:"#愤怒的毒奶"},
      {vid:"0x01",vname:"#毒奶集锦"},
      {vid:"0x02",vname:"#不忍直视"}
    ]
  },
  methods:{
    switchPart:function(vid){
      //fetch rsc
      var vsrc = document.querySelector('video');
      fetch('/video?vid='+vid)
      .then(function(res){
        return res.blob();
      })
      .then(function(myblob){
        var videoURL = URL.createObjectURL(myblob);
        if(flvjs.isSupported()){
          var flvrsc = flvjs.createPlayer({
            type:'flv',
            url:videoURL,
          });
          flvrsc.attachMediaElement(player);
          flvrsc.load();
        }
      });
      // console.log(vid)
    }
  },
})

new Vue({
  el:"#about",
  data:{},
  methods:{},
})

var chats = new Vue({
  el:"#screen",
  data:{
      contents:[

      ],
      ldProgress: 0,
      msg:"",
      qa: "middle",
      isplay:false,
  },

  methods:{

    playORpause:function(){

      if(player.paused){
        player.play();
        this.isplay = true;
      }else{
        player.pause();
        this.isplay = false;
      }
    },

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
})
;

noUiSlider.create(document.getElementById("slider"),{
  start:0,
  connect:[true,false],
  behaviour:'hover-snap',
  // direction: 'rtl',
  range:{
    'min':0,
    'max':100,
  },
})

//fetch rsc
var vsrc = document.querySelector('video');

fetch('/video?vid=0x00')
.then(function(res){
  return res.blob();
})
.then(function(myblob){

  var videoURL = URL.createObjectURL(myblob);
  // vsrc.src = videoURL ;
  if(flvjs.isSupported()){
    var flvrsc = flvjs.createPlayer({
      type:'flv',
      url:videoURL,
    });
    flvrsc.attachMediaElement(player);
    flvrsc.load();
  }
});


setInterval(function(){
  if (player.currentTime <= player.duration){
    //when downloading ....
    chats.ldProgress = player.buffered.end(0);

    var slider = document.getElementById("slider");
    slider.noUiSlider.on('change',function(){
       player.currentTime = this.get() / 100 * player.duration
    })
    slider.noUiSlider.set(player.currentTime/player.duration * 100);
  }
},100)


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
