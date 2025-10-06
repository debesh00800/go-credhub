class Solution {
    int lcmm(long long a,long long b){
        return (int)((a*b)/(__gcd(a,b)));
    }
public:
    vector<int> replaceNonCoprimes(vector<int>& nums) {
        int n=nums.size();
        stack<int> s;
        for(int i=0;i<n;i++){
            int x=nums[i];
            if(s.empty()){s.push(nums[i]);}
            else{
                while(!s.empty()){
                    if(__gcd(x,s.top())>1){
                        
                        x=lcmm(x,s.top());
                        s.pop();
                    }else{
                        break;
                    }
                }
                s.push(x);
            }
        }
        vector<int> ans;
        while(!s.empty()){
            ans.push_back(s.top());
            s.pop();
        }
        reverse(ans.begin(),ans.end());
        return ans;
    }
};
