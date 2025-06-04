export class ShareHostData {
    hostKey
    fileCount
}

export function bytesToSize(bytes) {
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    if (bytes === 0) return '0 B';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return (bytes / Math.pow(1024, i)).toFixed(2) + ' ' + sizes[i]; // toFixed(2)表示保留两位小数
}

export function timeStampDiff(shareId, timeStamp) {
    var ret = ""
    if (timeStamp == undefined) {
        return ret
    }
    if (timeStamp == 0) {
        ret = "永久有效"
        return ret
    }
    var diff = timeStamp - Math.floor(Date.now() / 1000)
    if (diff >= 0 && diff < 60) {
        ret = "即将过期"
    } else if (diff >= 60 && diff < 60 * 60) {
        ret = Math.floor(diff / 60) + "分钟"
    } else if (diff >= 60 * 60 && diff < 60 * 60 * 24) {
        ret = Math.floor(diff / (60 * 60)) + "小时"
    } else if (diff >= 60 * 60 * 24) {
        ret = Math.floor(diff / (60 * 60 * 24)) + "天"
    } else {
        ret = "已过期"
    }
    return ret
}

export function round(num, decimalPlaces) {
    const factor = Math.pow(10, decimalPlaces);
    return Math.round(num * factor) / factor;
}

export function shareDataExtend(row, isSelf) {
    var tmpDownLoadHistory = row
    tmpDownLoadHistory.expireTimeBtnColor = "success"
    if (row.share.expireTime == undefined) {
        tmpDownLoadHistory.expireTimeBtnTxt = "已过期"
        tmpDownLoadHistory.expireTimeBtnColor = "danger"
    } else if (row.share.expireTime == 0) {
        tmpDownLoadHistory.expireTimeBtnTxt = "永久有效"
    } else {
        var diff = row.share.expireTime - Math.floor(Date.now() / 1000)
        if (diff >= 0 && diff < 60) {
            tmpDownLoadHistory.expireTimeBtnTxt = "即将过期"
        } else if (diff >= 60 && diff < 60 * 60) {
            tmpDownLoadHistory.expireTimeBtnTxt = Math.floor(diff / 60) + "分钟"
        } else if (diff >= 60 * 60 && diff < 60 * 60 * 24) {
            tmpDownLoadHistory.expireTimeBtnTxt = Math.floor(diff / (60 * 60)) + "小时"
        } else if (diff >= 60 * 60 * 24) {
            tmpDownLoadHistory.expireTimeBtnTxt = Math.floor(diff / (60 * 60 * 24)) + "天"
        } else {
            tmpDownLoadHistory.expireTimeBtnTxt = "已过期"
            tmpDownLoadHistory.expireTimeBtnColor = "danger"
        }
    }
    if (tmpDownLoadHistory.expireTimeBtnTxt == "已过期" && (row.share.status == 0 || row.share.status == 2)) {
        tmpDownLoadHistory.downLoadBtnShow = false;
    } else {
        tmpDownLoadHistory.downLoadBtnShow = true;
    }

    tmpDownLoadHistory.downLoadTooltipStatus = "";
    if (row.status == 0) {
        tmpDownLoadHistory.downLoadBtnTxt = "下载";
        tmpDownLoadHistory.downLoadBtnShow = true;
        tmpDownLoadHistory.downLoadBtnEvent = "start";
        tmpDownLoadHistory.downLoadBtnLoad = false;
        tmpDownLoadHistory.openFileBtnShow = false;
        tmpDownLoadHistory.downLoadTooltipDisabled = false;
        tmpDownLoadHistory.downLoadTooltipPercentage = 0;
        tmpDownLoadHistory.downLoadTooltipTitle = "没有下载";
    } else if (row.status == 1) {
        tmpDownLoadHistory.downLoadBtnTxt = "取消";
        tmpDownLoadHistory.downLoadBtnShow = true;
        tmpDownLoadHistory.downLoadBtnEvent = "stop";
        tmpDownLoadHistory.downLoadBtnLoad = false;
        tmpDownLoadHistory.openFileBtnShow = false;
        tmpDownLoadHistory.downLoadTooltipDisabled = false;
        tmpDownLoadHistory.downLoadTooltipPercentage = parseFloat(((row.finish / row.size) * 100).toFixed(2));
        tmpDownLoadHistory.downLoadTooltipTitle = "等待下载";
    } else if (row.status == 2) {
        tmpDownLoadHistory.downLoadBtnTxt = "重新下载";
        tmpDownLoadHistory.downLoadBtnShow = false;
        tmpDownLoadHistory.downLoadBtnEvent = "start";
        tmpDownLoadHistory.downLoadBtnLoad = false;
        tmpDownLoadHistory.openFileBtnShow = true;
        tmpDownLoadHistory.downLoadTooltipDisabled = false;
        tmpDownLoadHistory.downLoadTooltipPercentage = parseFloat(((row.finish / row.size) * 100).toFixed(2));
        tmpDownLoadHistory.downLoadTooltipTitle = "完成下载";
        tmpDownLoadHistory.downLoadTooltipStatus = "success";
    } else if (row.status == 3) {
        tmpDownLoadHistory.downLoadBtnTxt = "停止";
        tmpDownLoadHistory.downLoadBtnShow = true;
        tmpDownLoadHistory.downLoadBtnEvent = "stop";
        tmpDownLoadHistory.downLoadBtnLoad = true;
        tmpDownLoadHistory.openFileBtnShow = false;
        tmpDownLoadHistory.downLoadTooltipDisabled = false;
        tmpDownLoadHistory.downLoadTooltipPercentage = parseFloat(((row.finish / row.size) * 100).toFixed(2));
        tmpDownLoadHistory.downLoadTooltipTitle = "下载中";
    } else if (row.status == 4) {
        tmpDownLoadHistory.downLoadBtnTxt = "继续下载";
        tmpDownLoadHistory.downLoadBtnShow = true;
        tmpDownLoadHistory.downLoadBtnEvent = "continue";
        tmpDownLoadHistory.downLoadBtnLoad = false;
        tmpDownLoadHistory.openFileBtnShow = false;
        tmpDownLoadHistory.downLoadTooltipDisabled = false;
        tmpDownLoadHistory.downLoadTooltipPercentage = parseFloat(((row.finish / row.size) * 100).toFixed(2));
        tmpDownLoadHistory.downLoadTooltipTitle = "已停止";
        tmpDownLoadHistory.downLoadTooltipStatus = "warning";
    } else if (row.status == 5 || row.status == 6 || row.status == 7) {
        tmpDownLoadHistory.downLoadBtnTxt = "重新下载";
        tmpDownLoadHistory.downLoadBtnShow = true;
        tmpDownLoadHistory.downLoadBtnEvent = "continue";
        tmpDownLoadHistory.downLoadBtnLoad = false;
        tmpDownLoadHistory.openFileBtnShow = false;
        tmpDownLoadHistory.downLoadTooltipDisabled = false;
        tmpDownLoadHistory.downLoadTooltipPercentage = parseFloat(((row.finish / row.size) * 100).toFixed(2));
        if (row.status == 5) {
            tmpDownLoadHistory.downLoadTooltipTitle = "下载出错";
        } else if (row.status == 6) {
            tmpDownLoadHistory.downLoadTooltipTitle = "共享文件已过期，无法下载";
        } else {
            tmpDownLoadHistory.downLoadTooltipTitle = "共享文件已删除，无法下载";
        }
        tmpDownLoadHistory.downLoadTooltipStatus = "exception";
    }
    if (tmpDownLoadHistory.share.shareId == 0) {
        tmpDownLoadHistory.downLoadBtnShow = false;
    }
    if (isSelf == 1) {
        tmpDownLoadHistory.downLoadBtnShow = false;
        tmpDownLoadHistory.downLoadTooltipDisabled = true;
        tmpDownLoadHistory.openFileBtnShow = true;
    }
    return tmpDownLoadHistory
}