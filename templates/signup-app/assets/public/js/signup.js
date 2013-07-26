var app = angular.module('Signup',[])
.directive('btnSubmit', function(){
  return function(scope, element, attrs){
    scope.$watch(function(){
      return scope.$eval(attrs.btnSubmit);
    },
    function(working){
      var el = $(element).button();
      if(working) el.button('loading');
      else el.button('reset');
    });
  }
});
Signup.$inject = ['$scope','$http'];
function Signup($scope, $http) {
  $scope.user = {};
  $scope.close = function() {
    $scope.error = '';
  }
  $http.get('signup.json').
    success(function(data){
      if(data.Ok){
        $scope.inviteOnly=true;
      }
    });
  $scope.update = function(user) {
    $scope.close();
    $scope.working = true;
    $scope.user = angular.copy(user);
    $http.post('signup.json', user).
      success(function(data){
        if(data.Ok){
          window.location.replace('signup-ok.html');
        } else {
          $scope.error = data.Error;
        }
        $scope.working = false;
      }).
      error(function(data,status){
        alert('HTTP '+status+': '+data);
        $scope.working = false;
      });
  }
}
