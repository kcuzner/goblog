/**
 * View for creating new feeds
 */
define(['_', 'ko', '$-extensions'], function (_, ko) {
	
	function NewFeedView() {
		
	}
	
	NewFeedView.prototype.newFeed = function () {
		var title = prompt('Enter a title for the new feed');
		if (title === null) {
			return;
		}
		if (!title) {
			alert('You must enter a title for the new feed.');
			return;
		}
		
		var path = prompt('Enter a path for the new feed', title.replace(/(^[\s#?=*&@]+)|([\s#?=*&@]+$)/, '').replace(/[\s#?=*&@]+/g, '-'));
		if (path === null) {
			return;
		}
		if (!path) {
			alert('You must enter a path for the new feed.');
			return;
		}
		
		if (path.indexOf('/') !== 0) {
			path = '/' + path;
		}
		
		var dto = {
			title: title,
			path: path
		};
		
		$.ajaxQ({
			url: '/feeds/save',
			type: 'POST',
			contentType: 'application/json',
			data: JSON.stringify(dto)
		}).then(function () {
			document.location.reload();
		}).fail(function (err) {
			alert('Unable to create new feed: ' + err);
			console && console.log(err);
		}).done();
	};
	
	return NewFeedView;
	
});
