app.controller('roomCtrl', function ($scope, $http, $websocket, $stateParams, $state) {
    var roomId = $stateParams.roomId

    $scope.roomStatus = null

    // when logged in status is broadcasted, update room status for this voter
    $scope.$on('logged-in', function() {
        updateVoterRoomStatus()
    });
    // attempt to update voter status right away, if voterId is available
    updateVoterRoomStatus()


    // retrieve room status for this room owner
    function updateVoterRoomStatus(){
        if (!$scope.voterId)
            return

        $http.get("api/room/"+roomId+"/voter/"+$scope.voterId).then(function (res){
            $scope.roomStatus = {happy:res.data.happy}
            startProcessingRoomEvents()
        },function(err){
            // room not found? redirect to error
            $state.go("index")
        })
    }


    // wrapper decorator for websocket response handler: parses incoming message
    // as json and passes to internal handler
    function jsonHandler(handler){
        return function(message){
            handler(JSON.parse(message.data))
        }
    }

    var processingEvents = false
    function startProcessingRoomEvents() {
        if (processingEvents)
            return
        processingEvents = true
        var ws = $websocket("ws://"+window.location.host+'/api/events/'+$stateParams.roomId)
        ws.onMessage(jsonHandler(function(roomInfo) {
            $scope.roomInfo = roomInfo

            $scope.title = roomInfo.name
            if (roomInfo.marks>0){
                $scope.title = '('+roomInfo.marks+') '+$scope.title
            }

            $scope.marks = new Array(roomInfo.marks)

            // something changed? maybe our own status on another window?
            updateVoterRoomStatus()
        }))
        $scope.$on("$destroy", function() {
            console.log("closing listener for room",$stateParams.roomId)
            ws.close()
        });
    }

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
