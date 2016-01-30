app.controller('newRoomCtrl', function ($scope, $http, $state) {
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
