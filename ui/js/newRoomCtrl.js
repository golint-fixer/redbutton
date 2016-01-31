app.controller('newRoomCtrl', function ($scope, $http, $state) {
    $scope.room = {name:""}

    $scope.setTitle("New Room")

    function setError(error){
        $scope.errorMessage = error
    }

    // form post: create room
    $scope.createRoom = function (){
        setError(null)
        $http.post("api/room", {name: $scope.room.name, owner: $scope.voterId},{headers:{'voter-id':$scope.voterId}}).then(createRoomCallback,errorCallback)
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
