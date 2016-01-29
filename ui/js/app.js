// wrapper decorator for websocket response handler: parses incoming message
// as json and passes to internal handler
function jsonHandler(handler){
    return function(message){
        handler(JSON.parse(message.data))
    }
}

angular.module('redButtonApp', [
  'angular-websocket',
  'ngCookies',
  'ui.router'
])
.config(function($stateProvider){
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
.controller('mainCtrl', function ($scope, $http, $cookies) {

    $scope.title="Loading..."

    $scope.voterId = null;
    function updateLogin(){
        $scope.voterId = $cookies.get("voterId")
        if ($scope.voterId)
            $scope.$broadcast("logged-in")
    }
    updateLogin()

    if (!$scope.voterId){
        $http.post("api/login").then(function(res){
            $cookies.put("voterId",res.data.voterId)
            updateLogin()
        })
    }

})


.controller('roomCtrl', function ($scope, $http, $websocket, $stateParams) {
    var roomId = $stateParams.roomId

    $scope.roomStatus = null

    // when logged in status is broadcasted, update room status for this voter
    $scope.$on('logged-in', function() {
        updateVoterRoomStatus()
    });


    // retrieve room status for this room owner
    function updateVoterRoomStatus(){
        if (!$scope.voterId)
            return

        $http.get("api/room/"+roomId+"/voter/"+$scope.voterId).then(function (res){
            $scope.roomStatus = {happy:res.data.happy}
        })
    }

    // start listening for room events
    $websocket("ws://"+window.location.host+'/api/events/'+$stateParams.roomId).onMessage(jsonHandler(function(roomInfo) {
        $scope.roomInfo = roomInfo

        $scope.title = roomInfo.name
        if (roomInfo.marks>0){
            $scope.title = '('+roomInfo.marks+') '+$scope.title
        }

        $scope.marks = new Array(roomInfo.marks)

        // something changed? maybe our own status on another window?
        updateVoterRoomStatus()
    }))

    function setHappy(happy){
        if ($scope.roomStatus.happy==happy)
            return;

        $http.post("api/room/"+roomId+"/voter/"+$scope.voterId,{happy:happy}).then(function (res){
            $scope.roomStatus = {happy:res.data.happy}
            $scope.roomStatus=res.data;
        })
    }

    // button handlers
    $scope.voteUp = function(){
        setHappy(true);
    }
    $scope.voteDown = function(){
        setHappy(false);
    }

})


.controller('newRoomCtrl', function ($scope, $http, $state) {
    $scope.room = {name:""}

    function setError(error){
        $scope.errorMessage = error
    }

    // form post: create room
    $scope.createRoom = function (){
        setError(null)
        $http.post("api/room", {name: $scope.room.name, owner: $scope.voterId}).then(createRoomCallback,errorCallback)
    }

    function createRoomCallback(res){
        var room = res.data
        $state.go("room-detail",{roomId:room.id})
    }

    function errorCallback(err) {
        console.log("error!",err)
        setError(err.data.message)
    }



})
