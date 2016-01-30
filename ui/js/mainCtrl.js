app.controller('mainCtrl', function ($scope, $http, $cookies) {

    $scope.title="Loading..."

    $scope.voterId = null;

    function updateLogin(){
        $scope.voterId = $cookies.get("voterId")
        if ($scope.voterId) {
            $scope.$broadcast("logged-in")
        }
    }
    updateLogin()

    if (!$scope.voterId){
        $http.post("api/login").then(function(res){
            $cookies.put("voterId",res.data.voterId)
            updateLogin()
        })
    }

    // a shortcut method for child scopes to update page title
    $scope.setTitle = function(title){
        $scope.title = title
    }

})