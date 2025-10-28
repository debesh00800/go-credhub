class Solution {
    string find(string s, string target1,int pos){
        string target=target1;
        map<char,int> m;
        map<char,int> m1;
        int n=target.size();
        for(int i=0;i<n;i++){
            m[s[i]]++;
            m1[s[i]]++;
        }
        for(auto it:m1){
            // if(pos==98)
            if(it.first>target[pos]){
                cout<<pos<<endl;
                target[pos]=it.first;
                m[it.first]--;
                if(m[it.first]==0){m.erase(it.first);}

                string t1="";
                for(int i=0;i<pos;i++){
                    t1+=target[i];
                }
                string t2="";
                for(int i=pos+1;i<n;i++){
                    t2+=target[i];
                }
                int flag=-1;
                for(int i=0;i<pos;i++){
                    if(m.find(t1[i])==m.end()){flag=1;
                        break;}
                    m[t1[i]]--;
                    if(m[t1[i]]==0){m.erase(t1[i]);}
                }
                if(flag==1){
                          for(auto it1:m){
                    m[it1.first]=0;
                }
                for(auto it1:m1){
                    m[it1.first]=it1.second;
                }
                
                    continue;}
                string ans2="";
                for(auto it:m){
                    int si=it.second;
                    char c=it.first;
                    for(int i=0;i<si;i++){
                        ans2+=c;
                    }
                }
                cout<<pos;
                string temp=t1+target[pos]+ans2;
                if(pos==98){cout<<temp<<endl;}
                if(temp.length()==n){return temp;}
                
                for(auto it1:m){
                    m[it1.first]=0;
                }
                for(auto it1:m1){
                    m[it1.first]=it1.second;
                }
            }
        }
        return "";
        
        
    }
public:
    string lexGreaterPermutation(string s, string target) {
        int n=target.length();
        string ans="";
        cout<<n-1<<endl;
        for(int i=n-1;i>=0;i--){
            ans=find(s,target,i);
            if(ans!=""){return ans;}
        }
        return ans;
    }
};
