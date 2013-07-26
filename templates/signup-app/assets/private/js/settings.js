var app = angular.module('Settings',[])
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
Settings.$inject = ['$scope','$http'];
function Settings($scope, $http) {
  $scope.user = {};
  $http.get('index.json').
    success(function(data){
      if(data.Ok){
        $scope.user = data;
      }
    });
  $scope.close = function() {
    $scope.error = '';
    $scope.saved = false;
  }
  $scope.update = function(user) {
    $scope.close();
    $scope.working = true;
    $scope.user = angular.copy(user);
    $http.post('settings.json', user).
      success(function(data){
        if(data.Ok){
          if(data.Changes > 0) {
            $scope.saved = true;
            $scope.user.OldPasswd = '';
            $scope.user.NewPasswd = '';
            $scope.user.Confirm = '';
          }
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
