/**
 * New post javascript
 */
define(['_', 'ko', 'ko-bindings'], function (_, ko) {
    function NewPostView() {
        var self = this;

        this.parsers = [{
            name: 'Markdown',
            mode: 'markdown'
        }, {
            name: 'HTML',
            mode: 'html'
        }];

        this.parser = ko.observable(this.parsers[0]);
        this.mode = ko.computed(function () {
            var parser = self.parser();
            return (parser && parser.mode) || '';
        });

        this.content = ko.observable();
    }

    return NewPostView;
});
