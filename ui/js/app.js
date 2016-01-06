// wrapper decorator for websocket response handler: parses incoming message
// as json and passes to internal handler
function jsonHandler(handler){
    return function(message){
        handler(JSON.parse(message.data))
    }
}

angular.module('redButtonApp', [
  'angular-websocket'
])
.controller('mainCtrl', function ($scope, $websocket) {

    $scope.title="Loading..."

    $scope.roomStatus = {
        happy: true
    }

    function setHappy(happy){
        if ($scope.roomStatus.happy==happy)
            return;

        $scope.roomStatus.happy=happy;
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
    }));
});
