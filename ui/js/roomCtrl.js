app.controller('roomCtrl', function ($scope, $http, $websocket, $stateParams, $state) {
    var roomId = $stateParams.roomId

    $scope.roomId = roomId
    $scope.setProjectorMode($stateParams['projector-mode']=='yes')

    $scope.roomInfo = null // updated after even from websocket
    $scope.voterStatus = null // updated after updateVoterRoomStatus() call


    $http.get("api/room/"+roomId).then(
    function (res){
        // room exists? start listening for room events
        startProcessingRoomEvents()
    },
    function (err){
        // no such room; fallback to index
        $state.go("index")
    })


    // retrieve room status for this room owner
    function updateVoterRoomStatus(){
        if (!$scope.voterId)
            return

        $http.get("api/room/"+roomId+"/voter/"+$scope.voterId).then(function (res){
            $scope.voterStatus=res.data;
        })
    }


    // wrapper decorator for websocket response handler: parses incoming message
    // as json and passes to internal handler
    function jsonHandler(handler){
        return function(message){
            handler(JSON.parse(message.data))
        }
    }

    function startProcessingRoomEvents() {
        processingEvents = true
        var ws = $websocket("ws://"+window.location.host+"/api/room/"+roomId+"/voter/"+$scope.voterId+"/events")

        // when controller is closed, close this websocket as well
        $scope.$on("$destroy", function() {
            ws.close()
        });

        // on message from json handler, update roomInfo with relevant data
        ws.onMessage(jsonHandler(function(roomInfo) {
            $scope.roomInfo = roomInfo

            $scope.setTitle(roomInfo.name)
            if (roomInfo.marks>0){
                $scope.setTitle('('+roomInfo.marks+') '+$scope.title)
            }

            $scope.marks = new Array(roomInfo.marks)

            // something changed? maybe our own status on another window?
            updateVoterRoomStatus()
        }))

    }

    $scope.setHappy = function (happy){
        if ($scope.voterStatus.happy==happy)
            return;

        $http.post("api/room/"+roomId+"/voter/"+$scope.voterId,{happy:happy}).then(function (res){
            $scope.voterStatus=res.data;
        })
    }

    $scope.resetHappy = function(){
        $http.post("api/room/"+roomId,{marks:0},{headers:{'voter-id':$scope.voterId}}).then(function (res){
            $scope.voterStatus=res.data;
        })
    }

})
