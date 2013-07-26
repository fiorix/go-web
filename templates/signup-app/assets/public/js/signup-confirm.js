var app = angular.module('SignupConfirm',[])
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
SignupConfirm.$inject = ['$scope','$http'];
function SignupConfirm($scope, $http) {
  $scope.user = {'URL':document.URL};
  $scope.close = function() {
    $scope.error = '';
  }
  $scope.update = function(user) {
    $scope.close();
    $scope.working = true;
    $scope.user = angular.copy(user);
    $http.post('../signup-confirm.json', user).
      success(function(data){
        if(data.Ok){
          window.location.replace('../u/');
        } else {
          $scope.error = data.Error;
        }
        $scope.working = false;
      }).
      error(function(data,status){
        if(status==404) {
          $scope.error = 'NotFound';
        } else {
          alert('HTTP '+status+': '+data);
        }
        $scope.working = false;
      });
  }
}
