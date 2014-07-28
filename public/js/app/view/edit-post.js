/**
 * New post javascript
 */
define(['_', 'ko', 'moment', 'model/post', 'ko-bindings'], function (_, ko, moment, Post) {

    function NewPostView(dto) {
        var self = this;

        this.working = ko.observable(0);

        this.feeds = dto.allFeeds;

        this.post = new Post(dto.post);

        if (dto.feeds.length) {
            //update the post's feeds to include these ones as a hint
            this.post.feeds(_(this.post.feeds()).union(dto.feeds).uniq().value());
        }
    }

    NewPostView.prototype.save = function() {
        var self = this;

        this.working(this.working() + 1);
        this.post.save()
            .fail(function (err) {
                alert('Unable to save post: ' + err);
                if (console && console.error) {
                    console.error(err);
                }
            })
            .fin(function () {
                self.working(self.working() - 1);
            })
            .done();
    };

    NewPostView.prototype.saveDraft = function() {
    };

    return NewPostView;
});
