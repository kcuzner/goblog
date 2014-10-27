/**
 * jQuery extensions
 */
define(['jquery', 'q'], function ($, Q) {

    /**
     * Q-promise ajax call
     * @param  {object} options Options for the ajax call
     * @return {Promise}         Jquery XHR interpreted as a promise
     */
    $.ajaxQ = function (options) {
        var deferred = Q.defer();

        $.ajax(options)
            .done(function (data) {
                deferred.resolve(data);
            })
            .fail(function (xhr, error, httperror) {
                var err = new Error(httperror || error);
                err.xhr = xhr;

                deferred.reject(err);
            });

        return deferred.promise;
    };

});
