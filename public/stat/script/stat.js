/**
 * Created by oboo_chin on 15/3/16.
 */

var model = {
    dataOutput: {},
    formHeader: {},
    classNames: {}
};


var Statistics = {
    /**
     * 初始化操作
     * @param {json} data 传入的原始数据
     * */
    init:function (data) {

        for (var itemName in data) {

            var itemData = {};
            itemData[itemName] = data[itemName];
            var itemClass = itemName.split(".")[1];

            if (Statistics.classCheck(itemClass)) {

                //添加新类并投入新类
                Statistics.addClass(itemClass);
                Statistics.dataDrop(itemClass, itemData);

            } else {
                //投入类
                Statistics.dataDrop(itemClass, itemData);
            }
        }

        Statistics.outputData(model.classNames);

    },

    /**
     * 输出数据
     * @param {json} classNames 传入重构json
     * */
    outputData: function (classNames) {

        var data = {dataList: classNames};

        var html = template('line', data);
        document.getElementById('content').innerHTML = html;
    },
    /**
     * 向数据栈加入数据
     * @param {json} className 传入数据栈json
     * @param {json} data 传入的数据栈
     * */
    dataDrop: function (className, data) {
        for (var name in data) {
            model.classNames[className][name] = data[name]
        }
    },
    /**
     * 检测class存在性
     * @param {json} className 类名
     * */
    classCheck: function (className) {

        if (model.classNames[className]) {
            return false;
        } else {
            return true;
        }
    },

    /**
     * 加入类名
     * @param {json} className 类名
     * */
    addClass: function (className) {
        model.classNames[className] = {};
    },
    tmp: {
        /**
         * 创建table
         * @param {string} name table名
         * @param {json} data 数据
         * */
        consTable: function (name, data) {
            var table = $("<table/>").attr("id", name);
            table.appendTo($("body"));
            table.id = name;
            for (var tr in data) {
                Statistics.tmp.consTr(tr, data[tr], table);
            }
        },
        /**
         * 创建table
         * @param {string} name tr名
         * @param {json} data 数据
         * @param {object} table table节点
         * */
        consTr: function (name, data, table) {
            var tr = $("<tr/>");
            $("<td/>").text(name).appendTo(tr);
            for (var item in data) {
                Statistics.tmp.formHeader[item] = 1;
                $("<td/>").text(data[item]).appendTo(tr);
            }
            tr.appendTo(table);
        },
        /**
         * 创建操作条
         * */
        consBar: function () {
            $("<div/>").attr("id", "jump").appendTo($("body"));
            for (var item in  model.classNames) {
                $("<a/>").text(item).attr("src", "#" + item).appendTo($("#jump"));
            }
        }
    }

};