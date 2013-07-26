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
  $scope.search = function(q) {
    $scope.working = true;
    $scope.q = angular.copy(q);
    $http.post('search.json', q).
      success(function(data){
        if(data.Ok){
          if(data.Results) {
            $scope.results = data.Results;
          } else {
            $scope.results = [];
          }
        }
        $scope.working = false;
      }).
      error(function(data,status){
        alert('HTTP '+status+': '+data);
        $scope.working = false;
      });
  }
}
