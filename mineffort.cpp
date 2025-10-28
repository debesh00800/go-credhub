#include <bits/stdc++.h>
using namespace std;

class Solution {
public:
    int rows, cols;
    vector<vector<int>> heights;
    int dirs[4][2] = {{1,0},{-1,0},{0,1},{0,-1}};

    bool dfs(int r, int c, int maxEffort, vector<vector<int>>& visited) {
        if (r == rows - 1 && c == cols - 1) return true;
        visited[r][c] = 1;

        for (auto &d : dirs) {
            int nr = r + d[0], nc = c + d[1];
            if (nr >= 0 && nr < rows && nc >= 0 && nc < cols && !visited[nr][nc]) {
                int diff = abs(heights[r][c] - heights[nr][nc]);
                if (diff <= maxEffort) {
                    if (dfs(nr, nc, maxEffort, visited)) return true;
                }
            }
        }
        return false;
    }

    int minimumEffortPath(vector<vector<int>>& h) {
        heights = h;
        rows = heights.size();
        cols = heights[0].size();

        int lo = 0, hi = 1e6, ans = 0;
        while (lo <= hi) {
            int mid = lo + (hi - lo) / 2;
            vector<vector<int>> visited(rows, vector<int>(cols, 0));
            if (dfs(0, 0, mid, visited)) {
                ans = mid;
                hi = mid - 1; // try smaller effort
            } else {
                lo = mid + 1; // need bigger effort
            }
        }
        return ans;
    }
};
