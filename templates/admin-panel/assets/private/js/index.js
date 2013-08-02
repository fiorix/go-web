var app = angular.module('Index',[])
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
Index.$inject = ['$scope','$http'];
function Index($scope, $http) {
  $scope.q = {};
  $scope.user = {};
  $http.get('index.json').
    success(function(data){
      if(data.Ok){
        $scope.user = {Email:data.Email};
      }
    });
}
