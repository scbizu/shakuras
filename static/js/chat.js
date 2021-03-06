Vue.use(VueMaterial)



var conn;

var player = document.getElementById("player");

//heiheihei
var user ="Alien"


function getCookie(name) {
  var value = "; " + document.cookie;
  var parts = value.split("; " + name + "=");
  if (parts.length == 2) return parts.pop().split(";").shift();
}

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
    getSeries:function(tag){
      fetch('/getSeries?tagname='+tag)
      .then(function(response){
        response.json().then(function(data){
          console.log(data)
          series.parts = data
          series.hasContent = true
        })
      })
      // console.log("clicked!")
      // sideBarRouter.push({path:"/watch/"+tagname})
    },
  }
})



var series = new Vue({
  el:"#series",
  data:{
    hasContent: false ,
    parts:[
      // {vid:"0x00",vname:"#愤怒的毒奶"},
      // {vid:"0x01",vname:"#毒奶集锦"},
      // {vid:"0x02",vname:"#不忍直视"}
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
        // if(flvjs.isSupported()){
        //   var flvrsc = flvjs.createPlayer({
        //     type:'flv',
        //     url:videoURL,
        //   });
        //   flvrsc.attachMediaElement(player);
        //   flvrsc.load();
        // }
        player.src = videoURL
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
      qa: "raw",
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
        conn.send(user+" :" + this.msg);
        this.msg ="";
    }
  },
})
;

var account = new Vue({
  el:"#account",
  data:{
    username:"",
    password:"",
    registClicked:false,
    loginClicked:false,
    vertical:"bottom",
    horizontal:"center",
    duration:2000,
    StatusOk:false,
    StatusInfo:"Failed",
    unlogin:true
  },
  methods:{
    auth: function(){
      if(document.cookie != ""){
        user = getCookie("username")
        this.unlogin = false
      }
    },

    login :function(){
      // this.loginClicked = true
      var payload = {
        username: this.username,
        password: this.password
      }

      fetch("/login",{
        method: "POST",
        headers:{
          "content-type":"application/json"
        },
        body: JSON.stringify(payload)
      })
      .then((res)=>{
        if (res.status == 200){
            console.log("success")
             user = this.username
             chats.contents.push("欢迎"+user+"阁下加入聊天室")
             this.unlogin = false
            //  this.loginClicked = false
        } else{
          res.json().then((data)=>{
            console.log(data)
            // this.$refs.snackbar.open();
            // this.loginClicked = false
          })
        }
      })
    },

    regist: function(){
          // this.registClicked = true
        var payload = {
          username: this.username,
          password: this.password
        }

        fetch('/regist',{
          method: "PUT",
          headers: {
            "content-type":"application/json"
          },
          body: JSON.stringify(payload)
        })
        .then((res)=>{

          if (res.status== 200) {
            res.json().then((data) =>{
                this.StatusOk = true
                this.StatusInfo = "Success"
                console.log(data)
                alert("registed")
            })

          }else{
            res.json().then((data) =>{
                  console.log(data)
                  alert(data)
            })
          }
        })
    }
  }
})

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

fetch('/firstvideo')
.then(function(res){
  return res.blob();
})
.then(function(myblob){

  var videoURL = URL.createObjectURL(myblob);
  // if(flvjs.isSupported()){
  //   var flvrsc = flvjs.createPlayer({
  //     type:'flv',
  //     url: videoURL,
  //   });
  //   flvrsc.attachMediaElement(player);
  //   flvrsc.load();
  // }
  player.src = videoURL
});



setInterval(function(){
  if (player.currentTime <= player.duration){
    chats.ldProgress = player.buffered.end(0)/player.duration * 100;

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
