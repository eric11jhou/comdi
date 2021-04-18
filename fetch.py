# encoding: utf-8
import urllib.request
import ssl
import json
import pdfplumber
from apscheduler.schedulers.blocking import BlockingScheduler
from datetime import datetime


def main():
    print(datetime.now().strftime("%Y-%m-%d %H:%M:%S"), '開始運行')
    ssl._create_default_https_context = ssl._create_unverified_context
    fetcher = Fetcher()
    fetcher.fetch()

    scheduler = BlockingScheduler()
    scheduler.add_job(new_fetch, 'cron', hour=16, minute=0)
    scheduler.start()


def new_fetch():
    fetcher = Fetcher()
    fetcher.fetch()


class Fetcher:
    def __init__(self):
        self.cur_key = ""
        self.all = {'Put': {}, 'Call': {}}

    def fetch(self):
        user_agent = "Mozilla/5.0 (Windows NT 6.3; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.95 Safari/537.36"
        headers = {'User-Agent': user_agent}
        req = urllib.request.Request(
            url='https://www.cmegroup.com/daily_bulletin/current/Section64_Metals_Option_Products.pdf', headers=headers)
        resp = urllib.request.urlopen(req)
        with open('Section64_Metals_Option_Products.pdf', 'wb') as f:
            f.write(resp.read())
        with pdfplumber.open(r'Section64_Metals_Option_Products.pdf') as pdf:
            for page in pdf.pages:
                page_data = page.extract_text()
                self.Put(page_data)
                self.Call(page_data)
        result = {'Time': datetime.now().strftime(
            "%Y-%m-%d %H:%M:%S"), 'Put': {}, 'Call': {}}
        for m in self.all['Put']:
            data = self.all['Put'][m]['Data']
            data.sort(key=lambda s: s[1], reverse=True)
            result['Put'] = {
                'Total': self.all['Put'][m]['Total'],
                'Rank': data[:6],
            }
            break
        for m in self.all['Call']:
            data = self.all['Call'][m]['Data']
            data.sort(key=lambda s: s[1], reverse=True)
            result['Call'] = {
                'Total': self.all['Call'][m]['Total'],
                'Rank': data[:6],
            }
            break
        try:
            with open('data.json', 'w', encoding='utf-8') as fs:
                json.dump(result, fs)
        except IOError as e:
            print(e)
        print(datetime.now().strftime("%Y-%m-%d %H:%M:%S"), '數據保存完成!')

    def Put(self, page_data):
        start = page_data.find('OG PUT COMEX GOLD OPTIONS')
        if start == -1:
            return
        end = page_data.find('OG CALL COMEX GOLD OPTIONS')
        if end == -1:
            end = page_data.find('THE INFORMATION')
        lines = page_data[start:end].splitlines()
        for line in lines:
            l = line.split()
            if len(l) > 10:
                if l[0].isdigit() == False:
                    continue
                if l[len(l)-1] == 'UNCH':
                    self.all['Put'][self.cur_key]['Data'].append(
                        (l[0], int(l[len(l)-2]), l[len(l)-1]))
                else:
                    self.all['Put'][self.cur_key]['Data'].append(
                        (l[0], int(l[len(l)-3]), l[len(l)-2]+l[len(l)-1]))
            else:
                if l[0][3:5].isdigit():
                    self.cur_key = l[0]
                    if self.cur_key not in self.all['Put']:
                        self.all['Put'][self.cur_key] = {
                            'Total': [], 'Data': []}
                elif l[0] == 'TOTAL':
                    total = l[len(l)-2]
                    change = l[len(l)-1]
                    if not total[len(total)-1].isdigit():
                        change = total[len(total)-1] + change
                        total = total[:len(total)-1]
                    self.all['Put'][self.cur_key]['Total'] = [total, change]

    def Call(self, page_data):
        start = page_data.find('OG CALL COMEX GOLD OPTIONS')
        if start == -1:
            return
        end = page_data.find('SILVER OPTIONS ON FUTURES')
        if end == -1:
            end = page_data.find('THE INFORMATION')
        lines = page_data[start:end].splitlines()
        for line in lines:
            l = line.split()
            if len(l) > 10:
                if l[0].isdigit() == False:
                    continue
                if l[len(l)-1] == 'UNCH':
                    self.all['Call'][self.cur_key]['Data'].append(
                        (l[0], int(l[len(l)-2]), l[len(l)-1]))
                else:
                    self.all['Call'][self.cur_key]['Data'].append(
                        (l[0], int(l[len(l)-3]), l[len(l)-2]+l[len(l)-1]))
            else:
                if l[0][3:5].isdigit():
                    self.cur_key = l[0]
                    if self.cur_key not in self.all['Call']:
                        self.all['Call'][self.cur_key] = {
                            'Total': [], 'Data': []}
                elif l[0] == 'TOTAL':
                    total = l[len(l)-2]
                    change = l[len(l)-1]
                    if not total[len(total)-1].isdigit():
                        change = total[len(total)-1] + change
                        total = total[:len(total)-1]
                    self.all['Call'][self.cur_key]['Total'] = [total, change]


if __name__ == '__main__':
    main()
