/**
 * User administration view
 */
 define(['_', 'ko', 'models/user', '$-extensions'], function (_, ko, User) {
    
    function UserAdminView() {
        var self = this;
        
        this.search = ko.observable();
        
        this.loading = ko.observable(0);
        this.results = ko.observable(null);
        this.selected = ko.observable();
        
        this.allRoles = allRoles; //defined in admin.amber
        
        var i = 0;
        this.search.subscribe(function (phrase) {
            var n = ++i;
            self.selected(null);
            self.results(null);
            
            self.loading(self.loading() + 1);
            $.ajaxQ({
                url: '/users/search',
                data: { phrase: phrase }
            }).then(function (dtos) {
                if (n === i) {
                    self.results(_.map(dtos, function (d) { return new User(d); }));   
                }
            }).fail(function (err) {
                alert('Unable to search: ' + err);
                console && console.log(err);
            }).fin(function () {
                self.loading(self.loading() - 1);
            }).done();
        })
    }
    
    UserAdminView.prototype.select = function (user) {
        this.selected(user);
    }
    
    return UserAdminView;
    
 })
 