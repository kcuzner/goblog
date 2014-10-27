/**
 * Knockout custom bindings
 */
define(['_', 'ko'], function (_, ko) {

    var instances_by_id = {} // needed for referencing instances during updates.
        , init_id = 0;       // generated id increment storage

    ko.bindingHandlers.ace = {
        init: function (element, valueAccessor) {
            var value = ko.utils.unwrapObservable(valueAccessor());

            // Ace attaches to the element by DOM id, so we need to make one for the element if it doesn't have one already.
            if (!element.id) {
                element.id = 'knockout-ace-' + init_id;
                init_id = init_id + 1;
            }

            var editor = ace.edit(element.id);

            editor.setValue(value);
            editor.gotoLine(0);

            editor.getSession().on("change", function (delta) {
                if (ko.isWriteableObservable(valueAccessor())) {
                    valueAccessor()(editor.getValue());
                }
            });

            instances_by_id[element.id] = editor;

            // destroy the editor instance when the element is removed
            ko.utils.domNodeDisposal.addDisposeCallback(element, function () {
                editor.destroy();
                delete instances_by_id[element.id];
            });
        },
        update: function (element, valueAccessor, allBindings, viewModel, bindingContext) {
            var value = ko.utils.unwrapObservable(valueAccessor());
            var id = element.id;

            var options = allBindings.get('aceOptions') || {};

            //handle programmatic updates to the observable
            // also makes sure it doesn't update it if it's the same.
            // otherwise, it will reload the instance, causing the cursor to jump.
            if (id !== undefined && id !== '' && instances_by_id.hasOwnProperty(id)) {
                var editor = instances_by_id[id];

                if (ko.utils.unwrapObservable(options.theme)) {
                    editor.setTheme("ace/theme/" + options.theme);
                }
                if (ko.utils.unwrapObservable(options.mode)) {
                    editor.getSession().setMode("ace/mode/" + options.mode);
                }
                if (ko.utils.unwrapObservable(options.readOnly)) {
                    editor.setReadOnly(true);
                }

                var content = editor.getValue();
                if (content !== value) {
                    editor.setValue(value);
                    editor.gotoLine(0);
                }
            }
        }
    };

    ko.aceEditors = {
        resizeAll: function () {
            for (var id in instances_by_id) {
                if (!instances_by_id.hasOwnProperty(id)) continue;
                var editor = instances_by_id[id];
                editor.resize();
            }
        },
        get: function (id) {
            return instances_by_id[id];
        }
    };

});
