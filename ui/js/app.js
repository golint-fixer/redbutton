app = angular.module('redButtonApp', [
  'angular-websocket',
  'ngCookies',
  'ui.router'
])

app.config(function($stateProvider){
  $stateProvider
    .state('index', {
      url: "",
      templateUrl: "pages/new-room.html",
      controller: "newRoomCtrl"
    })
    .state('room-detail', {
      url: "/room/:roomId",
      templateUrl: "pages/room.html",
      controller: "roomCtrl"
    })

})
