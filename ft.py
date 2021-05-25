import pandas as pd
import matplotlib.pyplot as plt
from collections import defaultdict
import numpy as np

seen = []

all = []

d = defaultdict(list)

dict_names = []

with open("timings.txt") as f:
    for line in f:
        lis = line.split()
        if len(lis) == 3:
            d[lis[1]].append(int(lis[2]))
            if lis[1] not in dict_names:
                dict_names.append(lis[1])

           # if int(lis[2]) < 5000:
            all.append(lis[2])

            #if int(lis[2]) > 2999:
            #    if lis[1] in seen:
            #        print("*** ", line)
            #    else:
            #        print(line)
            #    seen.append(lis[1])

#print(d)
averages = []

nof_buckets = 25

for i in dict_names:
    x = pd.Series(d[i])

    min = x.min()
    max = x.max()

    bucket_counts = [0] * (nof_buckets+1)
    bucket_width = (max-min)/nof_buckets

    for y in x:
        bucket = int((y-min)/bucket_width)
        bucket_counts[bucket] += 1

    # squash outlier buckets
    #for b in range(nof_buckets+1):
    #    if bucket_counts[b] <= 3:
    #        bucket_counts[b] = 0

    filtered_results = []
    for y in x:
        bucket = int((y-min)/bucket_width)
        if bucket_counts[bucket] > 0:
            filtered_results.append(y)

    x = pd.Series(filtered_results)

    # histogram on linear scale
    hist, bins, _ = plt.hist(x, bins=20)

    count = 0
    sum = 0
    for j in x:
        count += 1
        sum += j

    print(i)
    print(sum / count)
    averages.append(sum / count)
    plt.show()
