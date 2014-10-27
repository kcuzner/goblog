/**
 * Model for a post
 */
define(['_', 'ko', 'q', 'moment', '$-extensions'], function (_, ko, Q, moment) {
    
    function PostVersion(dto) {
        console.log(dto);
        this.created = moment(dto.created);
        this.path = dto.path;
        this.title = dto.title;
        this.content = dto.content;
        this.parser = dto.parser;
        this.tags = dto.tags;
    }
    
    PostVersion.prototype.toString = function () {
        return this.created.format('MMM D, YYYY hh:mm:ss A');
    }

    function Post(dto) {
        var self = this;

        this.id = dto.id;

        //array of feed dtos
        this.feeds = ko.observableArray();

        this.created = moment(dto.created);
        this.modified = moment(dto.modified);
        
        this.versions = ko.observable(_.map(dto.versions, function (d) { return new PostVersion(d); }));
        this.sortedVersions = ko.computed(function () {
            return _(self.versions())
                .sortBy(self.versions(), function (v) { return v.created.unix() })
                .reverse()
                .value();
        });
        var version = this.versions()[this.versions().length - 1] || {};

        this.title = ko.observable(version.title);
        this.path = ko.observable(version.path);

        this.parsers = [{
            name: 'Markdown',
            mode: 'markdown'
        }, {
            name: 'HTML',
            mode: 'html'
        }];

        this.parser = ko.observable(_.find(this.parsers, function (p) { return p.name === version.parser; }) || this.parsers[0]);
        this.mode = ko.computed(function () {
            var parser = self.parser();
            return (parser && parser.mode) || '';
        });

        this.content = ko.observable(version.content);

        this.tags = ko.observable((version.tags || []).join(' '));

        this.title.subscribe(function (t) {
            if (!self.path()) {
                var title = t.replace(/^[^a-zA-Z0-9]+/, '')
                    .replace(/[^a-zA-Z0-9]+$/, '')
                    .substring(0, 150)
                    .toLowerCase()
                    .replace(/[^a-zA-Z0-9]+/g, '-');
                self.path(self.created.format('/YYYY/MM/DD') + '/' + title);
            }
        });

        this.tags.subscribe(function (t) {
            self.tags(t.toLowerCase());
        });
    }

    /**
     * Transforms this post into an edit DTO
     * @return {object} Post edit dto
     */
    Post.prototype.toDTO = function() {
        return {
            id: this.id,
            feeds: this.feeds(),
            title: this.title(),
            path: this.path(),
            parser: this.parser().name,
            content: this.content(),
            tags: this.tags()
        };
    };
    
    /**
     * Updates a post from a DTO
     */
    Post.prototype.fromDTO = function (dto) {
        this.id = dto.id;
        this.feeds(dto.feeds);
        
        this.versions(_.map(dto.versions, function (d) { return new PostVersion(d); }));
        var version = this.versions()[this.versions().length - 1] || {};
        this.title(version.title);
        this.path(version.path);
        this.content(version.content);
        this.parser(_.find(this.parsers, function (p) { return p.name === version.parser; }) || this.parsers[0]);
        this.tags((version.tags || []).join(' '));
    }

    /**
     * Saves a post
     * @return {Promise} Will be fulfilled when saving complete
     */
    Post.prototype.save = function() {
        var self = this;
        var dto = this.toDTO();

        return $.ajaxQ({
            url: '/posts/edit',
            type: 'POST',
            contentType: 'application/json',
            data: JSON.stringify(dto)
        }).then(function (dto) {
            self.fromDTO(dto);
        });
    };

    return Post;

});
