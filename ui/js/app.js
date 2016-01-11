// wrapper decorator for websocket response handler: parses incoming message
// as json and passes to internal handler
function jsonHandler(handler){
    return function(message){
        handler(JSON.parse(message.data))
    }
}

angular.module('redButtonApp', [
  'angular-websocket', 'ngCookies'
])
.controller('mainCtrl', function ($scope, $http, $cookies, $websocket) {

    $scope.title="Loading..."

    $scope.roomStatus = null

    var voterId = null;
    function updateLogin(){
        voterId = $cookies.get("voterId")
        console.log("voterId", voterId)
        if (voterId)
            updateVoterStatus()
    }
    updateLogin()

    if (!voterId){
        console.log("logging in...")
        $http.post("login").then(function(res){
            console.log("login data:",res.data)
            $cookies.put("voterId",res.data.voterId)
            updateLogin()
        })
    }


    // retrieve room status for this room owner
    function updateVoterStatus(){
        if (!voterId)
            return

        console.log("updating voter status")
        $http.get("voter/"+voterId).then(function (res){
            console.log("voter status",res.data)
            $scope.roomStatus = {happy:res.data.happy}
        })
    }


    function setHappy(happy){
        if ($scope.roomStatus.happy==happy)
            return;

        $http.post("voter/"+voterId,{happy:happy}).then(function (res){
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



    // start listening for room events
    $websocket("ws://"+window.location.host+'/events').onMessage(jsonHandler(function(roomInfo) {
        $scope.roomInfo = roomInfo

        $scope.title = roomInfo.name
        if (roomInfo.marks>0){
            $scope.title = '('+roomInfo.marks+') '+$scope.title
        }

        $scope.marks = new Array(roomInfo.marks)

        // something changed? maybe our own status on another window?
        updateVoterStatus()
    }));



});
