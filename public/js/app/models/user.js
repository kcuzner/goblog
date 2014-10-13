/**
 * User model
 */
define(['_', 'ko', '$-extensions'], function (_, ko) {
    
    function User(dto) {
        var self = this;
        this.id = dto.id;
        this.username = ko.observable(dto.username);
        this.displayName = ko.observable(dto.displayName);
        this.roles = ko.observableArray(dto.roles);
        
        this.modified = ko.observable(false);
        var setModified = function () {
            self.modified(true);
        }
        this.username.subscribe(setModified);
        this.displayName.subscribe(setModified);
        this.roles.subscribe(setModified);
    }
    
    /**
     * Creates a save DTO for this user
     */
    User.prototype.toDTO = function () {
        return {
            id: this.id,
            username: this.username(),
            displayName: this.displayName(),
            roles: this.roles()
        }
    }
    
    User.prototype.save = function () {
        var self = this;
        var dto = this.toDTO();
        
        $.ajaxQ({
            url: '/users/save',
            type: 'PUT',
            contentType: 'application/json',
            data: JSON.stringify(dto)
        }).then(function () {
            self.modified(false);
        }).fail(function (err) {
            alert('Unable to save user: ' + err);
            console && console.log(err);
        }).done();
    }
    
    return User;
    
})
