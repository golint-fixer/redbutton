app = angular.module('redButtonApp', [
  'angular-websocket',
  'ngCookies',
  'ui.router'
])

app.config(function($stateProvider){
  $stateProvider
    .state('index', {
      url: "/",
      templateUrl: "pages/new-room.html",
      controller: "newRoomCtrl"
    })
    .state('room-detail', {
      url: "/room/:roomId?projector-mode",
      templateUrl: "pages/room.html",
      controller: "roomCtrl"
    })

})

app.run(['$state', function ($state) {
   $state.transitionTo('index');
}])