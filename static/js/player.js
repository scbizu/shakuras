var player = document.getElementById("player")

function playORpause(){
  if(player.paused){
    player.play();
  }else{
    player.pause();
  }
}

function fullScreen(){
    if (player.requestFullscreen) {
    player.requestFullscreen();
  } else if (player.mozRequestFullScreen) {
    player.mozRequestFullScreen();
  } else if (player.webkitRequestFullscreen) {
    player.webkitRequestFullscreen();
  }
}
