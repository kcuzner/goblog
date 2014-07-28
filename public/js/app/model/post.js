/**
 * Model for a post
 */
define(['_', 'ko', 'q', '$-extensions'], function (_, ko, Q) {

    function Post(dto) {
        var self = this;

        this.id = dto.id;

        //array of feed dtos
        this.feeds = ko.observableArray();

        this.created = moment(dto.created);
        this.modified = moment(dto.modified);
        this.title = ko.observable(dto.title);
        this.path = ko.observable(dto.path);

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

        this.title.subscribe(function (t) {
            if (!self.path()) {
                var title = t.substring(0, 150).toLowerCase().replace(/\s/g, '-');
                self.path(self.created.format('/YYYY/MM/DD') + '/' + title);
            }
        });
    }

    /**
     * Transforms this post into a DTO
     * @return {object} Post dto
     */
    Post.prototype.toDTO = function() {
        return {
            id: this.id,
            feeds: this.feeds(),
            title: this.title(),
            path: this.path(),
            parser: this.parser().name,
            content: this.content()
        };
    };

    /**
     * Saves a post
     * @return {Promise} Will be fulfilled when saving complete
     */
    Post.prototype.save = function() {
        var dto = this.toDTO();

        return $.ajaxQ({
            url: '/posts/edit',
            type: 'POST',
            contentType: 'application/json',
            data: JSON.stringify(dto)
        });
    };

    return Post;

});
